package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
	"github.com/ericogr/ads1115-to-mqtt/pkg/output"
	"github.com/ericogr/ads1115-to-mqtt/pkg/sensor"
)

const (
	// defaults
	DefaultServer      = "tcp://localhost:1883"
	DefaultClientID    = "ads1115-client"
	DefaultStateTopic  = "ads1115"
	perChannelTopicFmt = "ads1115/channel/%d"
	// discovery payload keys/values
	keyName                = "name"
	keyStateTopic          = "state_topic"
	keyUnitOfMeasurement   = "unit_of_measurement"
	keyDeviceClass         = "device_class"
	keyStateClass          = "state_class"
	keyValueTemplate       = "value_template"
	keyJSONAttributesTopic = "json_attributes_topic"
	keyUniqueID            = "unique_id"
	unitVolts              = "V"
	deviceClassVoltage     = "voltage"
	stateClassMeasurement  = "measurement"
	valueTemplateVoltage   = "{{ value_json.voltage }}"
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
		// per-channel discovery when discoveryTopic contains a formatter
		if strings.Contains(m.discoveryTopic, "%d") {
			for _, ch := range channels {
				if !ch.Enabled {
					continue
				}
				dTopic := fmt.Sprintf(m.discoveryTopic, ch.Channel)
				stateTopic := formatStateTopic(cfg.StateTopic, ch.Channel)
				name := discoveryName(cfg, &ch)
				uniqueID := discoveryUniqueID(cfg, &ch)
				payload := baseDiscoveryPayload(name, stateTopic, uniqueID)
				if err := publishJSON(client, dTopic, true, payload); err != nil {
					log.Printf("mqtt discovery publish error: %v", err)
				}
			}
		} else {
			name := discoveryName(cfg, nil)
			uniqueID := discoveryUniqueID(cfg, nil)
			payload := baseDiscoveryPayload(name, m.stateTopic, uniqueID)
			if err := publishJSON(client, m.discoveryTopic, true, payload); err != nil {
				log.Printf("mqtt discovery publish error: %v", err)
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
			topic = fmt.Sprintf(perChannelTopicFmt, r.Channel)
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

// helper: format a state topic for a channel using an optional formatter
func formatStateTopic(base string, ch int) string {
	if base != "" {
		if strings.Contains(base, "%d") {
			return fmt.Sprintf(base, ch)
		}
		return base
	}
	return fmt.Sprintf("ads1115/channel/%d", ch)
}

// helper: build a human-friendly discovery name; if ch != nil append channel
func discoveryName(cfg config.MQTTConfig, ch *config.ChannelConfig) string {
	name := cfg.DiscoveryName
	if name == "" {
		name = fmt.Sprintf("ADS1115 %s", cfg.ClientID)
	}
	if ch != nil {
		name = fmt.Sprintf("%s ch%d", name, ch.Channel)
	}
	return name
}

// helper: build a unique id for discovery; if ch != nil append channel
func discoveryUniqueID(cfg config.MQTTConfig, ch *config.ChannelConfig) string {
	uid := cfg.DiscoveryUniqueID
	if uid == "" {
		uid = cfg.ClientID
	}
	if uid != "" && ch != nil {
		uid = fmt.Sprintf("%s_%d", uid, ch.Channel)
	}
	return uid
}

// helper: base discovery payload map common to all entries
func baseDiscoveryPayload(name, stateTopic, uniqueID string) map[string]interface{} {
	payload := map[string]interface{}{
		keyName:                name,
		keyStateTopic:          stateTopic,
		keyUnitOfMeasurement:   unitVolts,
		keyDeviceClass:         deviceClassVoltage,
		keyStateClass:          stateClassMeasurement,
		keyValueTemplate:       valueTemplateVoltage,
		keyJSONAttributesTopic: stateTopic,
	}
	if uniqueID != "" {
		payload[keyUniqueID] = uniqueID
	}
	return payload
}

// helper: marshal and publish JSON payload
func publishJSON(client mqtt.Client, topic string, retained bool, payload map[string]interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token := client.Publish(topic, 0, retained, b)
	token.Wait()
	return token.Error()
}
