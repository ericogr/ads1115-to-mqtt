package sensor

import (
	"testing"
)

func TestConfigForChannelBytes(t *testing.T) {
	s := &ADS1115Sensor{}

	// channel 0, sample rate 128 -> expect msb 0xC3 lsb 0x83 (see implementation)
	msb, lsb, err := s.configForChannel(0, 128)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msb != 0xC3 || lsb != 0x83 {
		t.Fatalf("channel0@128 => got %02X %02X; want C3 83", msb, lsb)
	}

	// channel 1, sample rate 128 -> D3 83
	msb, lsb, err = s.configForChannel(1, 128)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msb != 0xD3 || lsb != 0x83 {
		t.Fatalf("channel1@128 => got %02X %02X; want D3 83", msb, lsb)
	}

	// sample rate 8 for channel 0 -> msb C3 lsb 03 (dr=0)
	msb, lsb, err = s.configForChannel(0, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msb != 0xC3 || lsb != 0x03 {
		t.Fatalf("channel0@8 => got %02X %02X; want C3 03", msb, lsb)
	}

	// invalid channel
	_, _, err = s.configForChannel(9, 128)
	if err == nil {
		t.Fatalf("expected error for invalid channel")
	}
}
