package sensor

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
)

type FakeSensor struct {
	channels []int
	scale    float64
	offset   float64
	mu       sync.Mutex
}

func NewFakeSensor(cfg config.Config) (Sensor, error) {
	return &FakeSensor{channels: cfg.Channels, scale: cfg.CalibrationScale, offset: cfg.CalibrationOffset}, nil
}

func (f *FakeSensor) Read() ([]Reading, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	out := make([]Reading, 0, len(f.channels))
	for _, ch := range f.channels {
		raw := int16(rand.Intn(32767))
		// simulate voltage in range 0..4.096
		value := float64(raw)/32767.0*4.096*f.scale + f.offset
		out = append(out, Reading{Channel: ch, Raw: raw, Value: value, Timestamp: now})
	}
	return out, nil
}

func (f *FakeSensor) Close() error { return nil }
