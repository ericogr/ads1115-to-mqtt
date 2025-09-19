package config

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalConfigJSON(t *testing.T) {
	js := `{
        "i2c": { "bus": "2", "address": 72 },
        "sample_rate": 128,
        "outputs": [{"type":"console"}],
        "sensor_type":"real",
        "channels": [
            {"channel": 0, "enabled": true, "calibration_scale": 1.0, "calibration_offset": 0.12},
            {"channel": 1, "enabled": false, "calibration_scale": 0.98, "calibration_offset": -0.05}
        ]
    }`

	var cfg Config
	if err := json.Unmarshal([]byte(js), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if cfg.I2C.Address != 72 {
		t.Fatalf("i2c address: got %d", cfg.I2C.Address)
	}
	if cfg.SampleRate != 128 {
		t.Fatalf("sample_rate: got %d", cfg.SampleRate)
	}
	if cfg.SensorType != "real" {
		t.Fatalf("sensor_type: got %q", cfg.SensorType)
	}
	if len(cfg.Outputs) != 1 || cfg.Outputs[0].Type != "console" {
		t.Fatalf("outputs: %+v", cfg.Outputs)
	}
	if len(cfg.Channels) != 2 {
		t.Fatalf("channels len: %d", len(cfg.Channels))
	}
	if cfg.Channels[0].Channel != 0 || !cfg.Channels[0].Enabled || cfg.Channels[0].CalibrationOffset != 0.12 {
		t.Fatalf("channel0 incorrect: %+v", cfg.Channels[0])
	}
	if cfg.Channels[1].Channel != 1 || cfg.Channels[1].Enabled || cfg.Channels[1].CalibrationScale != 0.98 {
		t.Fatalf("channel1 incorrect: %+v", cfg.Channels[1])
	}
}
