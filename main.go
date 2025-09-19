package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ericogr/ads1115-to-mqtt/pkg/config"
	"github.com/ericogr/ads1115-to-mqtt/pkg/output"
	console "github.com/ericogr/ads1115-to-mqtt/pkg/output/console"
	mqttout "github.com/ericogr/ads1115-to-mqtt/pkg/output/mqtt"
	"github.com/ericogr/ads1115-to-mqtt/pkg/sensor"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	outs, err := initOutputs(cfg)
	if err != nil {
		log.Fatalf("outputs: %v", err)
	}

	s, err := initSensor(cfg)
	if err != nil {
		log.Fatalf("sensor init: %v", err)
	}
	defer s.Close()

	runLoop(cfg, s, outs)
}

// loadConfig loads configuration from flags or a JSON file.
func loadConfig() (config.Config, error) {
	return config.LoadFromFlags()
}

// initOutputs constructs the configured outputs (console, mqtt, ...).
type outputEntry struct {
	Out        output.Output
	IntervalMs int
}

func initOutputs(cfg config.Config) ([]outputEntry, error) {
	entries := make([]outputEntry, 0, len(cfg.Outputs))
	for _, o := range cfg.Outputs {
		typ := strings.ToLower(o.Type)
		interval := o.IntervalMs
		if interval == 0 {
			interval = cfg.IntervalMs
		}
		switch typ {
		case "console":
			entries = append(entries, outputEntry{Out: console.NewConsole(), IntervalMs: interval})
		case "mqtt":
			var mqttCfg config.MQTTConfig
			if o.MQTT != nil {
				mqttCfg = *o.MQTT
			} else {
				mqttCfg = config.MQTTConfig{Server: "tcp://localhost:1883", ClientID: "ads1115-client", Topic: "ads1115"}
			}
			mo, err := mqttout.NewMQTT(mqttCfg)
			if err != nil {
				return nil, fmt.Errorf("mqtt init: %w", err)
			}
			entries = append(entries, outputEntry{Out: mo, IntervalMs: interval})
		default:
			log.Printf("warning: unknown output '%s', ignoring", o.Type)
		}
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no outputs configured")
	}
	return entries, nil
}

// initSensor creates a sensor implementation (real ADS1115 or fake simulator).
func initSensor(cfg config.Config) (sensor.Sensor, error) {
	switch strings.ToLower(cfg.SensorType) {
	case "simulation", "sim", "fake":
		return sensor.NewFakeSensor(cfg)
	default:
		return sensor.NewADS1115Sensor(cfg)
	}
}

// runLoop starts the periodic read/publish loop and handles shutdown.
func runLoop(cfg config.Config, s sensor.Sensor, outs []outputEntry) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// latest readings protected by mutex
	var mu sync.RWMutex
	var latest []sensor.Reading

	// sensor reader
	sensorTicker := time.NewTicker(time.Duration(cfg.IntervalMs) * time.Millisecond)
	defer sensorTicker.Stop()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-sensorTicker.C:
				readings, err := s.Read()
				if err != nil {
					log.Printf("read error: %v", err)
					continue
				}
				mu.Lock()
				latest = readings
				mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	// start output goroutines
	for _, e := range outs {
		entry := e
		go func() {
			ticker := time.NewTicker(time.Duration(entry.IntervalMs) * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					mu.RLock()
					snapshot := make([]sensor.Reading, len(latest))
					copy(snapshot, latest)
					mu.RUnlock()
					if len(snapshot) == 0 {
						continue
					}
					if err := entry.Out.Publish(snapshot); err != nil {
						log.Printf("output publish error: %v", err)
					}
				case <-done:
					return
				}
			}
		}()
	}

	log.Printf("started; sensor_type=%s outputs=%v interval=%dms", cfg.SensorType, cfg.Outputs, cfg.IntervalMs)

	// show effective configuration at startup
	if b, err := json.MarshalIndent(cfg, "", "  "); err == nil {
		fmt.Printf("config:\n%s\n", string(b))
	}

	<-stop
	close(done)
	log.Println("shutting down")
	for _, e := range outs {
		_ = e.Out.Close()
	}
}
