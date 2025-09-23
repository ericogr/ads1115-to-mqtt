package console

import (
	"fmt"
	"time"

	"github.com/ericogr/ads1115-to-mqtt/pkg/output"
	"github.com/ericogr/ads1115-to-mqtt/pkg/sensor"
)

type ConsoleOutput struct{}

func NewConsole() output.Output { return &ConsoleOutput{} }

func (c *ConsoleOutput) Publish(readings []sensor.Reading) error {
	for _, r := range readings {
		fmt.Printf("%s channel=%d raw=%d value=%.6f\n", r.Timestamp.Format(time.RFC3339), r.Channel, r.Raw, r.Value)
	}
	return nil
}

func (c *ConsoleOutput) Close() error { return nil }
