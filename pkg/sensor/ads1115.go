package sensor

import (
	"fmt"
	"time"

	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	pointerConv   = 0x00
	pointerConfig = 0x01
)

type ADS1115Sensor struct {
	dev        *i2c.Dev
	bus        i2c.BusCloser
	channels   []int
	sampleRate int
	scale      float64
	offset     float64
	pgaFS      float64
}

func NewADS1115Sensor(cfg config.Config) (Sensor, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("host init: %w", err)
	}
	bus, err := i2creg.Open(cfg.I2CBus)
	if err != nil {
		return nil, fmt.Errorf("open i2c: %w", err)
	}
	dev := &i2c.Dev{Addr: uint16(cfg.I2CAddress), Bus: bus}
	return &ADS1115Sensor{dev: dev, bus: bus, channels: cfg.Channels, sampleRate: cfg.SampleRate, scale: cfg.CalibrationScale, offset: cfg.CalibrationOffset, pgaFS: 4.096}, nil
}

func (s *ADS1115Sensor) Close() error {
	if s.bus != nil {
		return s.bus.Close()
	}
	return nil
}

func (s *ADS1115Sensor) Read() ([]Reading, error) {
	out := make([]Reading, 0, len(s.channels))
	now := time.Now()
	for _, ch := range s.channels {
		msb, lsb, err := s.configForChannel(ch)
		if err != nil {
			return nil, err
		}
		// write config
		if err := s.dev.Tx([]byte{pointerConfig, msb, lsb}, nil); err != nil {
			return nil, fmt.Errorf("write config: %w", err)
		}
		// wait for conversion (simple sleep)
		delayMs := int(1000.0/float64(s.sampleRate)) + 2
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		// read conversion
		readBuf := make([]byte, 2)
		if err := s.dev.Tx([]byte{pointerConv}, readBuf); err != nil {
			return nil, fmt.Errorf("read conv: %w", err)
		}
		raw := int16(readBuf[0])<<8 | int16(readBuf[1])
		value := float64(raw)*s.pgaFS/32768.0*s.scale + s.offset
		out = append(out, Reading{Channel: ch, Raw: raw, Value: value, Timestamp: now})
	}
	return out, nil
}

func (s *ADS1115Sensor) configForChannel(channel int) (byte, byte, error) {
	var mux byte
	switch channel {
	case 0:
		mux = 0x4
	case 1:
		mux = 0x5
	case 2:
		mux = 0x6
	case 3:
		mux = 0x7
	default:
		return 0, 0, fmt.Errorf("invalid channel %d", channel)
	}
	// PGA: use ±4.096V -> bits 001
	pga := byte(0x1)
	// data rate bits
	var dr byte
	switch s.sampleRate {
	case 8:
		dr = 0x0
	case 16:
		dr = 0x1
	case 32:
		dr = 0x2
	case 64:
		dr = 0x3
	case 128:
		dr = 0x4
	case 250:
		dr = 0x5
	case 475:
		dr = 0x6
	case 860:
		dr = 0x7
	default:
		dr = 0x4
	}
	var config uint16 = 0x8000 // OS = 1 (start single conversion)
	config |= uint16(mux) << 12
	config |= uint16(pga) << 9
	config |= 1 << 8 // single-shot mode
	config |= uint16(dr) << 5
	// comparator default: disabled (bits 1:0 = 11)
	config |= 0x3
	return byte(config >> 8), byte(config & 0xFF), nil
}
