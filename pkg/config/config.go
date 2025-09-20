package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MQTTConfig struct {
    Server   string `json:"server"`
    Username string `json:"username"`
    Password string `json:"password"`
    ClientID string `json:"client_id"`
    // StateTopic is the MQTT topic where the sensor state/value is published.
    StateTopic string `json:"state_topic"`
    // DiscoveryTopic is the full topic to publish Home Assistant discovery payload to
    // (for example: `homeassistant/sensor/machine_battery/config`). If empty, discovery is not published.
    DiscoveryTopic string `json:"discovery_topic,omitempty"`
    // Optional discovery payload fields for Home Assistant
    DiscoveryName     string `json:"discovery_name,omitempty"`
    DiscoveryUniqueID string `json:"discovery_unique_id,omitempty"`
}

type OutputConfig struct {
	Type       string      `json:"type"`
	IntervalMs int         `json:"interval_ms,omitempty"`
	MQTT       *MQTTConfig `json:"mqtt,omitempty"`
}

// ChannelConfig holds per-channel parameters: enabled, calibration and optional sample rate.
type ChannelConfig struct {
	Channel           int     `json:"channel"`
	Enabled           bool    `json:"enabled"`
	SampleRate        int     `json:"sample_rate,omitempty"`
	CalibrationScale  float64 `json:"calibration_scale,omitempty"`
	CalibrationOffset float64 `json:"calibration_offset"`
}

type I2CConfig struct {
	Bus     string `json:"bus"`
	Address int    `json:"address"`
}

type Config struct {
	I2C        I2CConfig       `json:"i2c"`
	SampleRate int             `json:"sample_rate"`
	Outputs    []OutputConfig  `json:"outputs"`
	SensorType string          `json:"sensor_type"`
	Channels   []ChannelConfig `json:"channels"`
}

func DefaultConfig() Config {
	return Config{
		I2C:        I2CConfig{Bus: "2", Address: 0x48},
		SampleRate: 128,
		Outputs:    []OutputConfig{{Type: "console"}},
		SensorType: "real",
		Channels: []ChannelConfig{
			{Channel: 0, Enabled: false, CalibrationScale: 1.0, CalibrationOffset: 0.0},
			{Channel: 1, Enabled: false, CalibrationScale: 1.0, CalibrationOffset: 0.0},
			{Channel: 2, Enabled: false, CalibrationScale: 1.0, CalibrationOffset: 0.0},
			{Channel: 3, Enabled: false, CalibrationScale: 1.0, CalibrationOffset: 0.0},
		},
	}
}

// LoadFromFlags loads configuration from a JSON file (optional) and flags.
// Flags override values present in the JSON file.
func LoadFromFlags() (Config, error) {
	cfgPath := flag.String("config", "", "Path to JSON config file")
	flagI2CBus := flag.String("i2c-bus", "", "I2C bus (e.g., '2' -> /dev/i2c-2)")
	flagI2CAddStr := flag.String("i2c-address", "", "I2C address (decimal or 0x hex)")
	flagSampleRate := flag.Int("sample-rate", -1, "ADS1115 sample rate (SPS)")
	flagOutputs := flag.String("outputs", "", "Comma-separated outputs (console,mqtt)")
	flagOutputIntervals := flag.String("output-intervals", "", "Comma-separated output intervals e.g. console=1000,mqtt=5000")
	flagMQTTServer := flag.String("mqtt-server", "", "MQTT server (tcp://host:port)")
	flagMQTTUser := flag.String("mqtt-user", "", "MQTT username")
	flagMQTTPass := flag.String("mqtt-pass", "", "MQTT password")
	flagSensorType := flag.String("sensor-type", "", "sensor type: real|simulation")
	flagChannelScales := flag.String("channel-scales", "", "Comma-separated per-channel scales e.g. 0=1.0,1=0.98")
	flagChannelOffsets := flag.String("channel-offsets", "", "Comma-separated per-channel offsets e.g. 0=0.12,1=-0.05")
	flagChannelSampleRates := flag.String("channel-sample-rates", "", "Comma-separated per-channel sample rates e.g. 0=250,1=128")
	flagChannelEnabled := flag.String("channel-enabled", "", "Comma-separated per-channel enabled flags e.g. 0=true,1=false")
    flagClientID := flag.String("mqtt-client-id", "", "MQTT client id")
    flagStateTopic := flag.String("mqtt-state-topic", "", "MQTT state topic to publish readings (e.g. sensors/machine_battery/voltage)")
    flagDiscoveryTopic := flag.String("mqtt-discovery-topic", "", "MQTT topic to publish Home Assistant discovery payload (full topic)")
    flagDiscoveryName := flag.String("mqtt-discovery-name", "", "Discovery: sensor name")
    flagDiscoveryUniqueID := flag.String("mqtt-discovery-unique-id", "", "Discovery: unique_id")

	flag.Parse()

	cfg := DefaultConfig()

	// If no -config was provided, try to load ./config.json by default
	if *cfgPath == "" {
		if _, err := os.Stat("config.json"); err == nil {
			*cfgPath = "config.json"
		}
	}

	if *cfgPath != "" {
		b, err := os.ReadFile(*cfgPath)
		if err != nil {
			return cfg, fmt.Errorf("read config: %w", err)
		}
		// Unmarshal into the new Config shape (channels are per-channel objects).
		if err := json.Unmarshal(b, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config: %w", err)
		}
	}

	if *flagI2CBus != "" {
		cfg.I2C.Bus = *flagI2CBus
	}
	if *flagI2CAddStr != "" {
		v, err := parseIntOrHex(*flagI2CAddStr)
		if err != nil {
			return cfg, fmt.Errorf("i2c-address: %w", err)
		}
		cfg.I2C.Address = v
	}
	if *flagSampleRate != -1 {
		cfg.SampleRate = *flagSampleRate
	}
	// Note: per-channel mappings (scales/offsets) are handled below; global calibration flags were removed in favor of per-channel flags.
	if *flagOutputs != "" {
		// convert simple CSV of types into structured OutputConfig entries
		parts := parseCSV(*flagOutputs)
		outs := make([]OutputConfig, 0, len(parts))
		for _, p := range parts {
			outs = append(outs, OutputConfig{Type: p})
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
    if *flagMQTTServer != "" || *flagMQTTUser != "" || *flagMQTTPass != "" || *flagClientID != "" || *flagStateTopic != "" || *flagDiscoveryTopic != "" {
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
                if *flagStateTopic != "" {
                    cfg.Outputs[i].MQTT.StateTopic = *flagStateTopic
                }
                if *flagDiscoveryTopic != "" {
                    cfg.Outputs[i].MQTT.DiscoveryTopic = *flagDiscoveryTopic
                }
                if *flagDiscoveryName != "" {
                    cfg.Outputs[i].MQTT.DiscoveryName = *flagDiscoveryName
                }
                if *flagDiscoveryUniqueID != "" {
                    cfg.Outputs[i].MQTT.DiscoveryUniqueID = *flagDiscoveryUniqueID
                }
                applied = true
            }
        }
        if !applied {
			// create mqtt output config and apply flags
			mqttOut := OutputConfig{Type: "mqtt", MQTT: &MQTTConfig{}}
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
                if *flagStateTopic != "" {
                    mqttOut.MQTT.StateTopic = *flagStateTopic
                }
                if *flagDiscoveryTopic != "" {
                    mqttOut.MQTT.DiscoveryTopic = *flagDiscoveryTopic
                }
                if *flagDiscoveryName != "" {
                    mqttOut.MQTT.DiscoveryName = *flagDiscoveryName
                }
                if *flagDiscoveryUniqueID != "" {
                    mqttOut.MQTT.DiscoveryUniqueID = *flagDiscoveryUniqueID
                }
                cfg.Outputs = append(cfg.Outputs, mqttOut)
            }
        }
	if *flagSensorType != "" {
		cfg.SensorType = *flagSensorType
	}
	// channel enabling is handled via -channel-enabled mapping

	// per-channel mappings via flags (override file values)
	if *flagChannelEnabled != "" {
		m, err := parseKeyBoolMap(*flagChannelEnabled)
		if err != nil {
			return cfg, err
		}
		for ch, val := range m {
			applied := false
			for i := range cfg.Channels {
				if cfg.Channels[i].Channel == ch {
					cfg.Channels[i].Enabled = val
					applied = true
					break
				}
			}
			if !applied {
				cfg.Channels = append(cfg.Channels, ChannelConfig{Channel: ch, Enabled: val, CalibrationScale: 1.0, CalibrationOffset: 0.0})
			}
		}
	}

	if *flagChannelScales != "" {
		m, err := parseKeyFloatMap(*flagChannelScales)
		if err != nil {
			return cfg, err
		}
		for ch, v := range m {
			applied := false
			for i := range cfg.Channels {
				if cfg.Channels[i].Channel == ch {
					cfg.Channels[i].CalibrationScale = v
					applied = true
					break
				}
			}
			if !applied {
				cfg.Channels = append(cfg.Channels, ChannelConfig{Channel: ch, Enabled: false, CalibrationScale: v, CalibrationOffset: 0.0})
			}
		}
	}

	if *flagChannelOffsets != "" {
		m, err := parseKeyFloatMap(*flagChannelOffsets)
		if err != nil {
			return cfg, err
		}
		for ch, v := range m {
			applied := false
			for i := range cfg.Channels {
				if cfg.Channels[i].Channel == ch {
					cfg.Channels[i].CalibrationOffset = v
					applied = true
					break
				}
			}
			if !applied {
				cfg.Channels = append(cfg.Channels, ChannelConfig{Channel: ch, Enabled: false, CalibrationScale: 1.0, CalibrationOffset: v})
			}
		}
	}

	if *flagChannelSampleRates != "" {
		m, err := parseKeyIntMap(*flagChannelSampleRates)
		if err != nil {
			return cfg, err
		}
		for ch, v := range m {
			applied := false
			for i := range cfg.Channels {
				if cfg.Channels[i].Channel == ch {
					cfg.Channels[i].SampleRate = v
					applied = true
					break
				}
			}
			if !applied {
				cfg.Channels = append(cfg.Channels, ChannelConfig{Channel: ch, Enabled: false, SampleRate: v, CalibrationScale: 1.0, CalibrationOffset: 0.0})
			}
		}
	}
	// NOTE: outputs[].interval_ms defaulting and sensor interval calculation are handled in the caller (main) based on sample_rate and channels

	// validate sample rate
	allowed := []int{8, 16, 32, 64, 128, 250, 475, 860}
	valid := false
	for _, a := range allowed {
		if cfg.SampleRate == a {
			valid = true
			break
		}
	}
	if !valid {
		return cfg, fmt.Errorf("invalid sample_rate %d; allowed: %v", cfg.SampleRate, allowed)
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

// parseChannels removed: use -channel-enabled mapping instead of shorthand CSV.

func parseKeyFloatMap(s string) (map[int]float64, error) {
	out := map[int]float64{}
	parts := parseCSV(s)
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid mapping '%s'", p)
		}
		k := strings.TrimSpace(kv[0])
		vstr := strings.TrimSpace(kv[1])
		ki, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("invalid channel '%s': %w", k, err)
		}
		vf, err := strconv.ParseFloat(vstr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value for channel %d: %w", ki, err)
		}
		out[ki] = vf
	}
	return out, nil
}

func parseKeyIntMap(s string) (map[int]int, error) {
	out := map[int]int{}
	parts := parseCSV(s)
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid mapping '%s'", p)
		}
		k := strings.TrimSpace(kv[0])
		vstr := strings.TrimSpace(kv[1])
		ki, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("invalid channel '%s': %w", k, err)
		}
		vi, err := strconv.Atoi(vstr)
		if err != nil {
			return nil, fmt.Errorf("invalid int value for channel %d: %w", ki, err)
		}
		out[ki] = vi
	}
	return out, nil
}

func parseKeyBoolMap(s string) (map[int]bool, error) {
	out := map[int]bool{}
	parts := parseCSV(s)
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid mapping '%s'", p)
		}
		k := strings.TrimSpace(kv[0])
		vstr := strings.TrimSpace(kv[1])
		ki, err := strconv.Atoi(k)
		if err != nil {
			return nil, fmt.Errorf("invalid channel '%s': %w", k, err)
		}
		vb, err := strconv.ParseBool(vstr)
		if err != nil {
			return nil, fmt.Errorf("invalid bool value for channel %d: %w", ki, err)
		}
		out[ki] = vb
	}
	return out, nil
}
