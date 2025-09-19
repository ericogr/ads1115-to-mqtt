package sensor

import "time"

type Reading struct {
	Channel   int       `json:"channel"`
	Raw       int16     `json:"raw"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type Sensor interface {
	Read() ([]Reading, error)
	Close() error
}
