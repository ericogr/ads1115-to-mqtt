package mqtt

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
	"github.com/ericogr/ads1115-to-mqtt/pkg/output"
	"github.com/ericogr/ads1115-to-mqtt/pkg/sensor"
)

type MQTTOutput struct {
	client         mqtt.Client
	stateTopic     string
	discoveryTopic string
}

func NewMQTT(cfg config.MQTTConfig, channels []config.ChannelConfig) (output.Output, error) {
	opts := mqtt.NewClientOptions().AddBroker(cfg.Server).SetClientID(cfg.ClientID)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	st := cfg.StateTopic
	m := &MQTTOutput{client: client, stateTopic: st, discoveryTopic: cfg.DiscoveryTopic}

	// Publish Home Assistant discovery payload(s) if requested
	if m.discoveryTopic != "" {
		// If discoveryTopic contains %d, publish per-channel discovery entries.
		if strings.Contains(m.discoveryTopic, "%d") {
			for _, ch := range channels {
				if !ch.Enabled {
					continue
				}
				dTopic := fmt.Sprintf(cfg.DiscoveryTopic, ch.Channel)
				// determine state_topic for this channel
				var stateTopic string
				if cfg.StateTopic != "" {
					if strings.Contains(cfg.StateTopic, "%d") {
						stateTopic = fmt.Sprintf(cfg.StateTopic, ch.Channel)
					} else {
						stateTopic = cfg.StateTopic
					}
				} else {
					stateTopic = fmt.Sprintf("ads1115/channel/%d", ch.Channel)
				}
				// build discovery payload
				name := cfg.DiscoveryName
				if name == "" {
					name = fmt.Sprintf("ADS1115 %s", cfg.ClientID)
				}
				name = fmt.Sprintf("%s ch%d", name, ch.Channel)
				uniqueID := cfg.DiscoveryUniqueID
				if uniqueID == "" {
					uniqueID = cfg.ClientID
				}
				if uniqueID != "" {
					uniqueID = fmt.Sprintf("%s_%d", uniqueID, ch.Channel)
				}
				payload := map[string]interface{}{
					"name":                  name,
					"state_topic":           stateTopic,
					"unit_of_measurement":   "V",
					"device_class":          "voltage",
					"state_class":           "measurement",
					"value_template":        "{{ value_json.voltage }}",
					"json_attributes_topic": stateTopic,
				}
				if uniqueID != "" {
					payload["unique_id"] = uniqueID
				}
				if b, err := json.Marshal(payload); err == nil {
					token := client.Publish(dTopic, 0, true, b)
					token.Wait()
				}
			}
		} else {
			// single discovery entry (no per-channel formatting)
			name := cfg.DiscoveryName
			if name == "" {
				name = fmt.Sprintf("ADS1115 %s", cfg.ClientID)
			}
			uniqueID := cfg.DiscoveryUniqueID
			if uniqueID == "" {
				uniqueID = cfg.ClientID
			}
			payload := map[string]interface{}{
				"name":                  name,
				"state_topic":           m.stateTopic,
				"unit_of_measurement":   "V",
				"device_class":          "voltage",
				"state_class":           "measurement",
				"value_template":        "{{ value_json.voltage }}",
				"json_attributes_topic": m.stateTopic,
			}
			if uniqueID != "" {
				payload["unique_id"] = uniqueID
			}
			if b, err := json.Marshal(payload); err == nil {
				token := client.Publish(m.discoveryTopic, 0, true, b)
				token.Wait()
			}
		}
	}

	return m, nil
}

func (m *MQTTOutput) Publish(readings []sensor.Reading) error {
	for _, r := range readings {
		// determine topic: if stateTopic contains a %d formatter use it for channel
		topic := m.stateTopic
		if strings.Contains(topic, "%d") {
			topic = fmt.Sprintf(topic, r.Channel)
		}
		if topic == "" {
			// fallback to per-channel topic
			topic = fmt.Sprintf("ads1115/channel/%d", r.Channel)
		}

		// publish payload including averaged value and raw (raw is sent as integer)
		payload := map[string]interface{}{"voltage": r.Value, "raw": r.Raw}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		token := m.client.Publish(topic, 0, false, b)
		token.Wait()
		if token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

func (m *MQTTOutput) Close() error {
	if m.client != nil {
		m.client.Disconnect(250)
	}
	return nil
}

// PublishRaw publishes a raw payload to the given topic. The caller can set the
// retain flag which is useful for discovery messages.
func (m *MQTTOutput) PublishRaw(topic string, payload []byte, retained bool) error {
	if m.client == nil {
		return fmt.Errorf("mqtt client not connected")
	}
	token := m.client.Publish(topic, 0, retained, payload)
	token.Wait()
	return token.Error()
}
