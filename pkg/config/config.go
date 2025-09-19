package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type MQTTConfig struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"client_id"`
	Topic    string `json:"topic"`
}

type OutputConfig struct {
	Type       string      `json:"type"`
	IntervalMs int         `json:"interval_ms,omitempty"`
	MQTT       *MQTTConfig `json:"mqtt,omitempty"`
}

type Config struct {
	I2CBus            string         `json:"i2c_bus"`
	I2CAddress        int            `json:"i2c_address"`
	SampleRate        int            `json:"sample_rate"`
	CalibrationScale  float64        `json:"calibration_scale"`
	CalibrationOffset float64        `json:"calibration_offset"`
	Outputs           []OutputConfig `json:"outputs"`
	SensorType        string         `json:"sensor_type"`
	Channels          []int          `json:"channels"`
	IntervalMs        int            `json:"interval_ms"`
}

func DefaultConfig() Config {
	return Config{
		I2CBus:            "2",
		I2CAddress:        0x48,
		SampleRate:        128,
		CalibrationScale:  1.0,
		CalibrationOffset: 0.0,
		Outputs:           []OutputConfig{{Type: "console", IntervalMs: 1000}},
		SensorType:        "real",
		Channels:          []int{0, 1, 2, 3},
		IntervalMs:        1000,
	}
}

// LoadFromFlags loads configuration from a JSON file (optional) and flags.
// Flags override values present in the JSON file.
func LoadFromFlags() (Config, error) {
	cfgPath := flag.String("config", "", "Path to JSON config file")
	flagI2CBus := flag.String("i2c-bus", "", "I2C bus (e.g., '2' -> /dev/i2c-2)")
	flagI2CAddStr := flag.String("i2c-address", "", "I2C address (decimal or 0x hex)")
	flagSampleRate := flag.Int("sample-rate", -1, "ADS1115 sample rate (SPS)")
	flagCalibration := flag.Float64("calibration", math.NaN(), "Calibration scale factor (multiplier)")
	flagCalOffset := flag.Float64("calibration-offset", math.NaN(), "Calibration offset")
	flagOutputs := flag.String("outputs", "", "Comma-separated outputs (console,mqtt)")
	flagOutputIntervals := flag.String("output-intervals", "", "Comma-separated output intervals e.g. console=1000,mqtt=5000")
	flagMQTTServer := flag.String("mqtt-server", "", "MQTT server (tcp://host:port)")
	flagMQTTUser := flag.String("mqtt-user", "", "MQTT username")
	flagMQTTPass := flag.String("mqtt-pass", "", "MQTT password")
	flagSensorType := flag.String("sensor-type", "", "sensor type: real|simulation")
	flagChannels := flag.String("channels", "", "Comma-separated channels e.g. 0,1,2,3")
	flagInterval := flag.Int("interval-ms", -1, "Publish interval in ms")
	flagClientID := flag.String("mqtt-client-id", "", "MQTT client id")
	flagTopic := flag.String("mqtt-topic", "", "MQTT topic base")

	flag.Parse()

	cfg := DefaultConfig()

	if *cfgPath != "" {
		b, err := os.ReadFile(*cfgPath)
		if err != nil {
			return cfg, fmt.Errorf("read config: %w", err)
		}
		if err := json.Unmarshal(b, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config: %w", err)
		}
	}

	if *flagI2CBus != "" {
		cfg.I2CBus = *flagI2CBus
	}
	if *flagI2CAddStr != "" {
		v, err := parseIntOrHex(*flagI2CAddStr)
		if err != nil {
			return cfg, fmt.Errorf("i2c-address: %w", err)
		}
		cfg.I2CAddress = v
	}
	if *flagSampleRate != -1 {
		cfg.SampleRate = *flagSampleRate
	}
	if !math.IsNaN(*flagCalibration) {
		cfg.CalibrationScale = *flagCalibration
	}
	if !math.IsNaN(*flagCalOffset) {
		cfg.CalibrationOffset = *flagCalOffset
	}
	if *flagOutputs != "" {
		// convert simple CSV of types into structured OutputConfig entries
		parts := parseCSV(*flagOutputs)
		outs := make([]OutputConfig, 0, len(parts))
		for _, p := range parts {
			outs = append(outs, OutputConfig{Type: p, IntervalMs: cfg.IntervalMs})
		}
		cfg.Outputs = outs
	}
	// parse output intervals mapping
	outIntervals := map[string]int{}
	if *flagOutputIntervals != "" {
		parts := parseCSV(*flagOutputIntervals)
		for _, p := range parts {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) != 2 {
				continue
			}
			if v, err := strconv.Atoi(kv[1]); err == nil {
				outIntervals[strings.TrimSpace(kv[0])] = v
			}
		}
		// apply to existing outputs
		for i := range cfg.Outputs {
			if v, ok := outIntervals[cfg.Outputs[i].Type]; ok {
				cfg.Outputs[i].IntervalMs = v
			}
		}
	}
	// map mqtt flags into the first mqtt output (create if missing)
	if *flagMQTTServer != "" || *flagMQTTUser != "" || *flagMQTTPass != "" || *flagClientID != "" || *flagTopic != "" {
		// Apply MQTT flags to all mqtt outputs; if none exist, create one.
		applied := false
		for i := range cfg.Outputs {
			if strings.ToLower(cfg.Outputs[i].Type) == "mqtt" {
				if cfg.Outputs[i].MQTT == nil {
					cfg.Outputs[i].MQTT = &MQTTConfig{}
				}
				if *flagMQTTServer != "" {
					cfg.Outputs[i].MQTT.Server = *flagMQTTServer
				}
				if *flagMQTTUser != "" {
					cfg.Outputs[i].MQTT.Username = *flagMQTTUser
				}
				if *flagMQTTPass != "" {
					cfg.Outputs[i].MQTT.Password = *flagMQTTPass
				}
				if *flagClientID != "" {
					cfg.Outputs[i].MQTT.ClientID = *flagClientID
				}
				if *flagTopic != "" {
					cfg.Outputs[i].MQTT.Topic = *flagTopic
				}
				applied = true
			}
		}
		if !applied {
			// create mqtt output config and apply flags
			mqttOut := OutputConfig{Type: "mqtt", IntervalMs: cfg.IntervalMs, MQTT: &MQTTConfig{}}
			if *flagMQTTServer != "" {
				mqttOut.MQTT.Server = *flagMQTTServer
			}
			if *flagMQTTUser != "" {
				mqttOut.MQTT.Username = *flagMQTTUser
			}
			if *flagMQTTPass != "" {
				mqttOut.MQTT.Password = *flagMQTTPass
			}
			if *flagClientID != "" {
				mqttOut.MQTT.ClientID = *flagClientID
			}
			if *flagTopic != "" {
				mqttOut.MQTT.Topic = *flagTopic
			}
			cfg.Outputs = append(cfg.Outputs, mqttOut)
		}
	}
	if *flagSensorType != "" {
		cfg.SensorType = *flagSensorType
	}
	if *flagChannels != "" {
		chs, err := parseChannels(*flagChannels)
		if err != nil {
			return cfg, err
		}
		cfg.Channels = chs
	}
	if *flagInterval != -1 {
		cfg.IntervalMs = *flagInterval
	}
	// ensure outputs have interval default
	for i := range cfg.Outputs {
		if cfg.Outputs[i].IntervalMs == 0 {
			cfg.Outputs[i].IntervalMs = cfg.IntervalMs
		}
	}

	if cfg.SampleRate <= 0 {
		return cfg, errors.New("sample-rate must be > 0")
	}

	return cfg, nil
}

func parseIntOrHex(s string) (int, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		v, err := strconv.ParseInt(s[2:], 16, 0)
		return int(v), err
	}
	v, err := strconv.Atoi(s)
	return v, err
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parseChannels(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		v, err := strconv.Atoi(t)
		if err != nil {
			return nil, fmt.Errorf("invalid channel '%s': %w", t, err)
		}
		out = append(out, v)
	}
	return out, nil
}
