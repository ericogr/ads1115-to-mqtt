package console

import (
	"fmt"
	"github.com/ericogr/ads1115-to-mqtt/pkg/output"
	"github.com/ericogr/ads1115-to-mqtt/pkg/sensor"
)

type ConsoleOutput struct{}

func NewConsole() output.Output { return &ConsoleOutput{} }

func (c *ConsoleOutput) Publish(readings []sensor.Reading) error {
	for _, r := range readings {
		fmt.Printf("%s channel=%d raw=%d value=%.6f\n", r.Timestamp.Format("2006-01-02T15:04:05Z07:00"), r.Channel, r.Raw, r.Value)
	}
	return nil
}

func (c *ConsoleOutput) Close() error { return nil }
