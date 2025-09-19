package mqtt

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
	"github.com/ericogr/ads1115-to-mqtt/pkg/output"
	"github.com/ericogr/ads1115-to-mqtt/pkg/sensor"
)

type MQTTOutput struct {
	client    mqtt.Client
	topicBase string
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
	return &MQTTOutput{client: client, topicBase: cfg.Topic}, nil
}

func (m *MQTTOutput) Publish(readings []sensor.Reading) error {
	for _, r := range readings {
		topic := fmt.Sprintf("%s/channel/%d", m.topicBase, r.Channel)
		payload, err := json.Marshal(r)
		if err != nil {
			return err
		}
		token := m.client.Publish(topic, 0, false, payload)
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
