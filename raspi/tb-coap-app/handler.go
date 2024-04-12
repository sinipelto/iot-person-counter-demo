package main

import (
	"base"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/udp"
	"github.com/plgd-dev/go-coap/v3/udp/client"
)

// Global one shared connection instance
var con *client.Conn
var mcon mqtt.Client

// Queues for messages
var telemetry chan []byte

// Sensors
var sensors map[string]*base.Sensor
var sensor0 *base.Sensor
var sensor1 *base.Sensor

func createMqttClient() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AutoReconnect = true
	opts.ConnectRetry = true
	opts.ConnectTimeout = time.Second * 10
	opts.PingTimeout = time.Second * 10
	opts.ConnectRetryInterval = time.Second * 10
	opts.Order = false
	opts.Username = cfg.Server.AccessToken
	opts.Servers = make([]*url.URL, 1)
	srv, _ := url.Parse(cfg.Server.Host)
	opts.Servers[0] = srv
	return mqtt.NewClient(opts)
}

func connectMqtt() {
	var ontm bool

	mcon = createMqttClient()
	tk := mcon.Connect()

	for ctr := 0; RUNNING && ctr < 100; ctr++ {
		ontm = tk.WaitTimeout(time.Second * 5)
		if ontm {
			break
		}
		fmt.Println("ERROR: Could not connect to server: Timeout.")
		// panic("ERR_CLNT_TMOUT")
	}
	err := tk.Error()
	if err != nil {
		fmt.Println("ERROR: Could not connect to server:", err.Error())
		panic("ERR_CLNT_CONN_FAIL")
	}
}

func connectCoap() {
	if !RUNNING {
		return
	}

	var err error

	if con != nil {
		ctx, cancel := context.WithTimeout(con.Context(), time.Second*5)
		defer cancel()
		err = con.Ping(ctx)
		if err == nil {
			return
		}
		fmt.Println("Connection down. Re-establishing..")
	}

	// Dispose old connection
	con, err = nil, nil

	// 3 * 100 sec = 300 sec = 5min
	// E.g. router reboot 2-3 min
	for ctr := 0; RUNNING && ctr < 100; ctr++ {
		con, err = udp.Dial(cfg.Server.Host)
		if err != nil {
			fmt.Println("ERROR: Could not connect to server:", err.Error())
			time.Sleep(time.Second * 5)
		} else {
			break
		}
	}

	if err != nil {
		fmt.Println("FATAL: Retries exceeded or interrupted. Could not connect to server:", err.Error())
		panic("ERR_UDP_DIAL_SERVER:" + err.Error())
	}
}

// Prepare connection and queues ready for sending
func InitCloud(wg *sync.WaitGroup) {
	telemetry = make(chan []byte, cfg.QueueSize)
	switch cfg.Server.Protocol {
	case base.Mqtt:
		connectMqtt() // or PANIC
	case base.Coap:
		connectCoap() // or PANIC
	default:
		panic("ERR_CFG_UNKNOWN_SRV_PROTO")
	}
	wg.Add(1)
	go messageHandler(wg)
}

func messageHandler(wg *sync.WaitGroup) {
	defer func() {
		switch cfg.Server.Protocol {
		case base.Mqtt:
			mcon.Disconnect(3000)
			fmt.Println("MQTT client disconnected.")
		case base.Coap:
			con.Close()
			fmt.Println("CoaP client disconnected.")
		default:
			panic("ERR_UNKNOWN_SRV_PROTO")
		}
		wg.Done()
	}()
Loop:
	for RUNNING {
		// TODO: handle re-send with higher priority
		// put failed message first in queue => use FIFO?
		select {
		case payload := <-telemetry:
			var err error
			switch cfg.Server.Protocol {
			case base.Mqtt:
				t := mcon.Publish("v1/devices/me/telemetry", 1, true, payload)
				ontm := t.WaitTimeout(time.Second * 5)
				if !ontm {
					fmt.Println("ERROR: Failed to send telemetry message: Timeout")
					telemetry <- payload
					continue Loop
				}
				err = t.Error()
			case base.Coap:
				ctx, cancel := context.WithTimeout(con.Context(), time.Second*10)
				defer cancel()

				// TODO API PATH + TELEMETRYPATH TO CONFIG!
				// TODO: LONG SEND TIME ISSUE!!!! OVER 5 SECONDS PER MESSAGE ON WORST CASE - FIGURE OUT THE PROBLEM
				_, err = con.Post(ctx, "/api/v1/"+cfg.Server.AccessToken+"/telemetry", message.AppJSON, bytes.NewReader(payload))
			default:
				panic("ERR_UNKNOWN_SRV_PROTO")
			}

			// If fails, try to reconnect, and re-queue message
			if err != nil {
				fmt.Println("ERROR: Failed to send telemetry message:", err.Error())
				// Re-queue failed message
				if RUNNING {
					fmt.Println("Trying to reconnect..")
					telemetry <- payload
					if cfg.Server.Protocol == base.Coap {
						connectCoap() // or PANIC
					} else {
						time.Sleep(time.Second * 5)
					}
				} else {
					break Loop
				}
			}

			fmt.Println("Telemetry sent OK:", string(payload[:]))

			// dec, err := resp.ReadBody()
			// if err != nil {
			// 	panic("ERR_COAP_READ_BODY" + err.Error())
			// }

			// fmt.Printf("Telemetry Response: ID: %d: CODE: %+v MSG: %+v\n", resp.MessageID(), resp.Code(), string(dec))
		case <-interrupt:
			fmt.Println("Message handler stopping.")
			break Loop
		}
	}
}

func sendTelemetry(data any) {
	b, err := json.Marshal(data)

	if err != nil {
		panic("ERR_TELEMETRY_MARSHAL_JSON: " + err.Error())
	}

	// Works if neeeded, testing etc
	// _, err = con.AsyncPing(func() {
	// 	fmt.Println("Got PONG.")
	// })
	// if err != nil {
	// 	panic("ERR_CONF_PING_CB" + err.Error())
	// }

	telemetry <- b
}

// Asyncly handle the event
// Take COPY of the values, might change in the caller
// Transparent generic handler function to lift out stuff from the main loop
func HandleMessage(wg *sync.WaitGroup, raw *base.Message) {
	// Message Format:
	// : XXX_YYY(_EXTRA)

	// Trim newline chars from the end
	data := strings.TrimSpace(raw.Content)

	// If multiple messages send at once, process each
	msgs := strings.Split(data, base.LF_STR)

	for _, msg := range msgs {
		msg = strings.TrimSpace(msg)
		arr := strings.Split(msg, "_")

		// Expect Format:
		// MSGTYPE_SERIAL(_OPTIONAL)
		if len(arr) < 2 {
			fmt.Println("ERROR: Unexpected packet format received. Skipping.")
			continue
		}

		sensor, ok := sensors[arr[1]]

		if !ok {
			fmt.Println("ERROR: Detectedd device Serial Number not in config. Skipping.")
			continue
		}

		// Packet type
		switch arr[0] {
		case "M":
			// M_SERIAL
			if len(arr) != 2 {
				fmt.Println("ERROR: Unexpected packet format received. Skipping.")
				continue
			}
			fmt.Println("Motion packet.")
			go handleMotion(wg, sensor, &raw.Stamp)
		case "P":
			// P_SERIAL_COUNTER
			if len(arr) != 3 {
				fmt.Println("ERROR: Unexpected packet format received. Skipping.")
				continue
			}
			fmt.Println("Ping packet.")
			go handlePing(wg, sensor, arr[2], &raw.Stamp)
		default:
			fmt.Println("ERROR: Unknown message type received. Skipping.")
		}
	}

	wg.Done()
}

func handleMotion(wg *sync.WaitGroup, sensor *base.Sensor, tm *time.Time) {
	// Get timestamp as epoch int64 for cloud to understand
	ts := tm.Unix() * 1000

	// Form the msg packet
	pkt := &base.MotionPacket{Stamp: tm}

	msg := &base.MotionTelemetry{
		Ts: ts,
		Values: &base.MotionPacketWrapper{
			"MOTION_" + sensor.Config.Serial: pkt,
		},
	}

	// Send telemetry info to cloud immediately bg
	// Title: MOTION_UNO123
	go sendTelemetry(msg)

	// If current = sensor1
	// Check if sensor 0 has triggered within timeout => person--
	// If current = sensor0
	// Check if sensor 1 has triggered within timeout => person++

	// fmt.Printf("BT: S0: %+v S1: %+v", sensor0.LastMotion, sensor1.LastMotion)
	sensor.LastMotion = pkt
	// fmt.Printf("AT: S0: %+v S1: %+v", sensor0.LastMotion, sensor1.LastMotion)

	if sensor.Config.Index == 1 {
		// fmt.Printf("TS1: S0: %+v S1: %+v", sensor0.LastMotion, sensor1.LastMotion)
		if sensor0.LastMotion != nil && sensor0.LastMotion.Stamp.Add(time.Millisecond*time.Duration(sensor0.Config.TriggerTimeout)).After(*tm) {
			// fmt.Printf("TS1+TM: S0: %+v S1: %+v", sensor0.LastMotion, sensor1.LastMotion)
			fmt.Println("PERSON DELTA: -1")
			dp := &base.PersonDeltaTelementry{
				Ts: ts,
				Values: &base.PersonDeltaPacketWrapper{
					"PERSON_DELTA": -1,
				},
			}
			go sendTelemetry(dp)
		}
	} else {
		// fmt.Printf("TS2: S0: %+v S1: %+v", sensor0.LastMotion, sensor1.LastMotion)
		if sensor1.LastMotion != nil && sensor1.LastMotion.Stamp.Add(time.Millisecond*time.Duration(sensor1.Config.TriggerTimeout)).After(*tm) {
			// fmt.Printf("TS2+TM: S0: %+v S1: %+v", sensor0.LastMotion, sensor1.LastMotion)
			fmt.Println("PERSON DELTA: +1")
			dp := &base.PersonDeltaTelementry{
				Ts: ts,
				Values: &base.PersonDeltaPacketWrapper{
					"PERSON_DELTA": +1,
				},
			}
			go sendTelemetry(dp)
		}
	}

	// fmt.Printf("MOTION HANDLED: %+v\n", pkt)
}

func handlePing(wg *sync.WaitGroup, sensor *base.Sensor, count string, tm *time.Time) {
	// fmt.Println("STARTED")

	// 8 16 32 or 64, depending on config
	num, err := strconv.ParseUint(count, 10, int(cfg.BitSize))

	if err != nil {
		fmt.Println("ERROR: Could not parse counter value to int:", err.Error())
		return
	}

	// pkt := &base.PingPacket{
	// 	Stamp: tm,
	// 	Count: uint32(num),
	// }

	// Update sensor last ping
	// sensor.LastPing = pkt

	// Create a ping telemetry packet to send over proto to the cloud
	msg := &base.PingTelemetry{
		Ts: tm.Unix() * 1000,
		Values: &base.PingPacketWrapper{
			"PING_" + sensor.Config.Serial: uint32(num),
		},
	}

	// Send ping telemetry to cloud
	// Title: PING_UNO123
	// dont go => this is already go'd, avoid connection hangs, stack growth
	// sendTelemetry(&base.PingPacketWrapper{"PING_" + pkt.Serial: pkt})
	go sendTelemetry(msg)

	// fmt.Printf("PING HANDLED: %+v\n", pkt)
}
