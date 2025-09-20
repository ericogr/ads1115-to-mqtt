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

func NewMQTT(cfg config.MQTTConfig) (output.Output, error) {
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

	// Publish Home Assistant discovery payload if requested
	if m.discoveryTopic != "" {
		// build discovery payload using available fields
		name := cfg.DiscoveryName
		if name == "" {
			name = fmt.Sprintf("ADS1115 %s", cfg.ClientID)
		}
		uniqueID := cfg.DiscoveryUniqueID
		if uniqueID == "" {
			uniqueID = cfg.ClientID
		}
        payload := map[string]interface{}{
            "name":               name,
            "state_topic":        m.stateTopic,
            "unit_of_measurement": "V",
            "device_class":        "voltage",
        }
		if uniqueID != "" {
			payload["unique_id"] = uniqueID
		}
		if b, err := json.Marshal(payload); err == nil {
			token := client.Publish(m.discoveryTopic, 0, false, b)
			token.Wait()
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

		// publish simple payload expected by Home Assistant (e.g. { "voltage": 3.72 })
		payload := map[string]float64{"voltage": r.Value}
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
