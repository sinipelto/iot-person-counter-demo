package base

import (
	"net/url"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type Protocol uint8

const (
	Mqtt Protocol = iota
	Coap
)

func (p Protocol) String() string {
	switch p {
	case Mqtt:
		return "mqtt"
	case Coap:
		return "coap"
	default:
		panic("ERR_INVALID_PROTO")
	}
}

var StrToProtocol = map[string]Protocol{
	"mqtt":  Mqtt,
	"mqtts": Mqtt,
	"coap":  Coap,
	"coaps": Coap,
}

func GetProtocol(s string) Protocol {
	v, ok := StrToProtocol[strings.ToLower(s)]
	if !ok {
		panic("ERR_INVALID_PROTOCOL")
	}
	return v
}

type Data struct {
	Buffer []byte
	Length int
	Stamp  time.Time
}

type Message struct {
	Content string
	Stamp   time.Time
}

type SensorConfig struct {
	Serial         string
	Index          uint8
	TriggerTimeout uint32
}

type Server struct {
	Url         *url.URL
	AccessToken string
	Protocol    Protocol
}

type Serial struct {
	Port     string
	Baud     int
	Parity   serial.Parity
	StopBit  serial.StopBits
	DataBits byte
}

type Config struct {
	BufSize   uint8
	BitSize   uint8
	QueueSize uint8
	Serial    *Serial
	Server    *Server
	Sensors   []*SensorConfig
}

type Sensor struct {
	Config     *SensorConfig
	LastMotion *MotionPacket
	// LastPing   *PingPacket
}

type MotionPacket struct {
	Stamp *time.Time `json:"stamp"`
}

// type PingPacket struct {
// 	Stamp *time.Time `json:"stamp"`
// 	Count uint32     `json:"count"`
// }

// type PersonDeltaPacket struct {
// 	Delta int8       `json:"delta"`
// }

type PersonDeltaTelemetry struct {
	Ts     int64                     `json:"ts"`
	Values *PersonDeltaPacketWrapper `json:"values"`
}

type MotionTelemetry struct {
	Ts     int64                `json:"ts"`
	Values *MotionPacketWrapper `json:"values"`
}

type PingTelemetry struct {
	Ts     int64              `json:"ts"`
	Values *PingPacketWrapper `json:"values"`
}

// -1 / 0 / 1
type PersonDeltaPacketWrapper map[string]int8

// 0, 1, 2, 3....
type PingPacketWrapper map[string]uint32

type MotionPacketWrapper map[string]*MotionPacket
