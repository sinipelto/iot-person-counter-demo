package config

import (
	"base"
	"encoding/json"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/tarm/serial"
)

const DEFAULT_CONFIG_FILE = "config.json"

func ParseConfig(fn string) *base.Config {
	ex, err := os.Executable()
	if err != nil {
		panic("ERR_PATH_EXEC")
	}

	// /path/to/exec => /path/to/fn
	ex = filepath.Dir(ex)
	ex = filepath.Join(ex, fn)

	// If is symlink, process link
	stat, _ := os.Lstat(ex)
	if stat.Mode()&os.ModeSymlink != 0 {
		ex, err = os.Readlink(ex)
		if err != nil {
			panic("ERR_READ_LINK")
		}
	}

	f, err := os.Open(ex)
	if err != nil {
		panic("ERR_OPEN_CFG_FILE")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		panic("ERR_READ_CFG_FILE")
	}

	if !json.Valid(data) {
		panic("ERR_INVALID_JSON")
	}

	raw := &Config{}
	err = json.Unmarshal(data, raw)

	if err != nil {
		panic("ERR_CFG_UMRSL" + err.Error())
	}

	if raw.Serial == nil {
		panic("ERR_CFG_SRL_NLPTR")
	}

	if raw.Server == nil {
		panic("ERR_CFG_SRV_NLPTR")
	}

	if raw.Sensors == nil {
		panic("ERR_CFG_SNRS_NLPTR")
	}

	v := reflect.ValueOf(raw).Elem()
	// Loop struct fields, check if any nil
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i)
		if val.IsNil() {
			panic("ERR_FIELD_NULL_OR_UNDEFINED: " + v.Type().Field(i).Name)
		}
	}

	v = reflect.ValueOf(raw.Serial).Elem()
	// Loop struct fields, check if any nil
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i)
		if val.IsNil() {
			panic("ERR_FIELD_NULL_OR_UNDEFINED: " + v.Type().Field(i).Name)
		}
	}

	v = reflect.ValueOf(raw.Server).Elem()
	// Loop struct fields, check if any nil
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i)
		if val.IsNil() {
			panic("ERR_CFG_FIELD_NULL_OR_UNDEFINED: " + v.Type().Field(i).Name)
		}
	}

	if *raw.Serial.Baud < math.MinInt || *raw.Serial.Baud > math.MaxInt {
		panic("ERR_BAUD_OVERFLOW")
	}

	if *raw.Serial.DataBits < 0 || *raw.Serial.DataBits > math.MaxUint8 {
		panic("ERR_DATABITS_OVERFLOW")
	}

	if *raw.BufSize < 0 || *raw.BufSize > math.MaxUint8 {
		panic("ERR_BUFSIZE_OOR")
	}

	if *raw.BitSize < 0 || *raw.BitSize > math.MaxUint8 {
		panic("ERR_CFG_BITSIZE_OOR")
	}

	if *raw.BitSize < 0 || *raw.BitSize > math.MaxUint8 {
		panic("ERR_CFG_QSIZE_OOR")
	}

	if strings.TrimSpace(*raw.Server.Url) == "" {
		panic("ERR_CFG_EMPTY_SRV_URL")
	}

	if strings.TrimSpace(*raw.Server.AccessToken) == "" {
		panic("ERR_CFG_EMPTY_SRV_ACC_TOK")
	}

	surl, err := url.Parse(*raw.Server.Url)
	if err != nil {
		panic("ERR_CANT_PARSE_SRV_URL")
	}

	cfg := &base.Config{
		Serial: &base.Serial{
			Port:     *raw.Serial.Port,
			Baud:     int(*raw.Serial.Baud),
			DataBits: byte(*raw.Serial.DataBits),
		},
		BufSize:   uint8(*raw.BufSize),
		BitSize:   uint8(*raw.BitSize),
		QueueSize: uint8(*raw.QueueSize),
		Server: &base.Server{
			Url:         surl,
			AccessToken: *raw.Server.AccessToken,
			Protocol:    base.GetProtocol(surl.Scheme),
		},
	}

	switch *raw.Serial.Parity {
	case "none":
		cfg.Serial.Parity = serial.ParityNone
	default:
		panic("ERR_CFG_PARITY_UNDEFINED")
	}

	switch *raw.Serial.StopBit {
	case "1":
		cfg.Serial.StopBit = serial.Stop1
	case "1half", "15", "1_5", "1.5":
		cfg.Serial.StopBit = serial.Stop1Half
	case "2":
		cfg.Serial.StopBit = serial.Stop2
	default:
		panic("ERR_CFG_STOPBIT_UNDEFINED")
	}

	if len(*raw.Sensors) <= 0 {
		panic("ERR_CFG_NO_SENSORS_CONFIGURED")
	}

	for _, e := range *raw.Sensors {
		v := reflect.ValueOf(e).Elem()
		// Loop struct fields, check if any nil
		for i := 0; i < v.NumField(); i++ {
			val := v.Field(i)
			if val.IsNil() {
				panic("ERR_CFG_FIELD_NULL_OR_UNDEFINED: " + v.Type().Field(i).Name)
			}
		}
	}

	for _, e := range *raw.Sensors {
		if *e.Index < 0 || *e.Index > math.MaxUint8 {
			panic("ERR_CFG_SENSOR_INDEX_OOR")
		}
		if *e.TriggerTimeout < 0 || *e.TriggerTimeout > math.MaxUint32 {
			panic("ERR_CFG_SENSOR_TMOUT_OOR")
		}
		cfg.Sensors = append(cfg.Sensors, &base.SensorConfig{
			Serial:         *e.Serial,
			Index:          uint8(*e.Index),
			TriggerTimeout: uint32(*e.TriggerTimeout),
		})
	}

	return cfg
}
