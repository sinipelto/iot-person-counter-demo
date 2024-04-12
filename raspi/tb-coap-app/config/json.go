package config

// Generated by https://quicktype.io

type Config struct {
	BufSize   *int64     `json:"bufSize"`
	BitSize   *int64     `json:"bitSize"`
	QueueSize *int64     `json:"queueSize"`
	Serial    *Serial    `json:"serial"`
	Sensors   *[]*Sensor `json:"sensors"`
	Server    *Server    `json:"server"`
}

type Sensor struct {
	Serial         *string `json:"serial"`
	Index          *int64  `json:"index"`
	TriggerTimeout *int64  `json:"triggerTimeout"`
}

type Serial struct {
	Port     *string `json:"port"`
	Baud     *int64  `json:"baud"`
	Parity   *string `json:"parity"`
	StopBit  *string `json:"stopBit"`
	DataBits *int64  `json:"dataBits"`
}

type Server struct {
	Host        *string `json:"host"`
	Protocol    *string `json:"protocol"`
	AccessToken *string `json:"accessToken"`
}
