package output

import "github.com/ericogr/ads1115-to-mqtt/pkg/sensor"

type Output interface {
	Publish([]sensor.Reading) error
	Close() error
}

// helper constructors are in subpackages
