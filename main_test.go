package main

import (
	"testing"

	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
)

func TestComputeSensorInterval(t *testing.T) {
	// no enabled channels -> fallback to global sample rate
	cfg := config.Config{SampleRate: 128}
	if got := computeSensorInterval(cfg); got != 10 {
		t.Fatalf("fallback interval: got %d want 10", got)
	}

	// one enabled channel (default sample rate 128)
	cfg.Channels = []config.ChannelConfig{{Channel: 0, Enabled: true}}
	if got := computeSensorInterval(cfg); got != 10 {
		t.Fatalf("one channel interval: got %d want 10", got)
	}

	// two enabled channels at 128 -> ~20ms
	cfg.Channels = []config.ChannelConfig{{Channel: 0, Enabled: true}, {Channel: 1, Enabled: true}}
	if got := computeSensorInterval(cfg); got != 20 {
		t.Fatalf("two channel interval: got %d want 20", got)
	}

	// mixed sample rates: 128 and 250 -> expect 10 + 6 = 16
	cfg.Channels = []config.ChannelConfig{{Channel: 0, Enabled: true, SampleRate: 128}, {Channel: 1, Enabled: true, SampleRate: 250}}
	if got := computeSensorInterval(cfg); got != 16 {
		t.Fatalf("mixed interval: got %d want 16", got)
	}
}
