package sensor

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
)

type FakeSensor struct {
	channels       []int
	channelScales  map[int]float64
	channelOffsets map[int]float64
	mu             sync.Mutex
}

func NewFakeSensor(cfg config.Config) (Sensor, error) {
	chans, scales, offs, _ := buildChannelSettings(cfg)
	return &FakeSensor{channels: chans, channelScales: scales, channelOffsets: offs}, nil
}

func (f *FakeSensor) Read() ([]Reading, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	out := make([]Reading, 0, len(f.channels))
	for _, ch := range f.channels {
		raw := int16(rand.Intn(32767))
		// simulate voltage in range 0..4.096
		scale := 1.0
		off := 0.0
		if v, ok := f.channelScales[ch]; ok {
			scale = v
		}
		if v, ok := f.channelOffsets[ch]; ok {
			off = v
		}
		value := float64(raw)/32767.0*4.096*scale + off
		out = append(out, Reading{Channel: ch, Raw: raw, Value: value, Timestamp: now})
	}
	return out, nil
}

func (f *FakeSensor) Close() error { return nil }
