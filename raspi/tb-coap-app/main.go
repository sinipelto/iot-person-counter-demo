package main

import (
	"base"
	"config"
	"strings"
	"sync"

	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/tarm/serial"
)

var RUNNING bool
var cfg *base.Config
var interrupt chan os.Signal

func flushAndClose(p *serial.Port) {
	if p == nil {
		return
	}

	defer func() {
		recover()
		// if r := recover(); r != nil {
		// 	fmt.Printf("Recover panic: %+v\n", r)
		// }
	}()

	_ = p.Flush()
	// time.Sleep(time.Millisecond * 100)
	_ = p.Close()
	// time.Sleep(time.Millisecond * 100)

	fmt.Println("Serial port closed.")
}

func openOrAbort(conf *serial.Config, old *serial.Port) *serial.Port {
	// Ensure old flushed closed first
	if old != nil {
		flushAndClose(old)
	}

	prt, err := serial.OpenPort(conf)

	if err != nil {
		flushAndClose(prt)
		panic("FATAL_SERIAL_OPEN_FAIL:" + err.Error())
	}

	return prt
}

func main() {
	fmt.Println("APPLICATION START")

	RUNNING = true

	// Buffered, dont block sending
	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func(ch <-chan os.Signal) {
		<-ch
		RUNNING = false
	}(interrupt)

	cfg = config.ParseConfig(config.DEFAULT_CONFIG_FILE)

	// Init sensor objects
	sensors = make(map[string]*base.Sensor)
	for _, s := range cfg.Sensors {
		ss := &base.Sensor{Config: s}
		if s.Index == 0 {
			sensor0 = ss
		} else {
			sensor1 = ss
		}
		sensors[s.Serial] = ss
	}

	serialCfg := &serial.Config{
		Name:     cfg.Serial.Port,     // COM Port
		Baud:     cfg.Serial.Baud,     // Speed
		Size:     cfg.Serial.DataBits, // Data bits
		Parity:   cfg.Serial.Parity,   // Parity
		StopBits: cfg.Serial.StopBit,  // Stop bit 1,half,2
	}

	// Buf should handle max
	// PING(4)_(1)Serial(6bytes)_Counter(32bits/4bytes) = 4 + 1 + 6 + 1 + 8 = 20 bytes
	// buf := make([]byte, cfg.BufSize)

	var data string
	var err error

	// Buffered, dont block on send
	dataQ := make(chan *base.Data, cfg.QueueSize)

	wg := &sync.WaitGroup{}

	// Open the serial port and defer flush and close it on exit
	prt := openOrAbort(serialCfg, nil)
	defer flushAndClose(prt)
	time.Sleep(time.Millisecond * 500)

	// Init cloud connection
	// before port flush
	// To avoid serial stacking if connection fails
	InitCloud(wg)

	// Flush any old dirt first
	err = prt.Flush()

	// Should not fail
	if err != nil {
		panic("ERR_FLUSH_FAILED:" + err.Error())
	}

	// Start reading in bg to avoid block
	// Might block for long time, dont wait if interrupted
	go func(ch chan<- *base.Data) {
		for RUNNING {
			chunk := &base.Data{Buffer: make([]byte, cfg.BufSize)}
			chunk.Length, err = prt.Read(chunk.Buffer)
			// Collect timestamp for complete message as soon as possible after recv
			// Arduino has no RTC so we need to use Raspi RTC/system time
			chunk.Stamp = time.Now()
			if err != nil {
				fmt.Println("ERROR: Error reading serial:", err.Error())
				if RUNNING {
					prt = openOrAbort(serialCfg, prt)
				}
			} else {
				ch <- chunk
			}
		}
	}(dataQ)

MainLoop:
	for RUNNING {
		// Wait for either single serial read
		// or os interrupt to stop exec
		select {
		case chunk := <-dataQ:
			// fmt.Println("Got read.")

			// len read > 0
			// if n <= 0 {
			// 	fmt.Println("No data received.")
			// 	continue
			// }

			// collect input byte[] => string
			// Keep track of buffered input from sub-buffer
			data += string(chunk.Buffer[:chunk.Length])
			// nt += n

			// fmt.Printf("Read %d bytes.\n", n)
			// fmt.Printf("Total %d bytes.\n", nt)

			// fmt.Println("RAW:", buf[:cfg.BufSize])
			// fmt.Println("DATA:", "'", data, "'")

			// Wait for message terminator
			// Otherwise serial not finished
			// Process all messages collected so far
			for bf, af, found := strings.Cut(data, base.LF_STR); found; bf, af, found = strings.Cut(data, base.LF_STR) {
				msg := &base.Message{Content: bf, Stamp: chunk.Stamp}

				// fmt.Println("MSG:", "'", msg.Content, "'")
				// fmt.Println("TIME:", msg.Stamp.String())

				// Handle received message
				wg.Add(1)
				go HandleMessage(wg, msg)

				// Clear handled message from buffer
				data = af
			}
		case <-interrupt:
			// fmt.Println("Got interrupt.")
			break MainLoop
		}

		// data = ""
		// nt = 0
	}

	// Wait for all calls to finish
	wg.Wait()
	fmt.Println("APPLICATION EXIT")
}
