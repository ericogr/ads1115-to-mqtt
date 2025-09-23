# ADS1115 to MQTT

Analog readings from an ADS1115 (4 channels) with configurable outputs (console, MQTT).

## Overview

- Read ADS1115 over I2C (or use a simulated sensor for testing).
- Configure one or more outputs via `outputs[]` in the config; each output has a `type` and optional `interval_ms`.
- Load configuration from `config.json` in the current directory by default, or pass `-config /path/to/file`. Command-line flags override file values.

## Main use cases

- Development / testing: `make run` (starts with a fake sensor and console output).
- Production: `make build` and run the produced binary with `-config` or flags.

## Quickstart

### Developer run (console + simulation)

```
make run
```

### Build

```
make build
```

### Run built binary (example)

```
./bin/ads1115-to-mqtt -outputs console,mqtt -output-intervals console=1000,mqtt=5000 -mqtt-server tcp://broker:1883
```

### Run container (example)

```
docker run --rm -it --name ads1115-to-mqtt -v "$(pwd)/config.json":/config.json:ro ericogr/ads1115-to-mqtt:latest -config /config.json
```

## Configuration

### Example config (`config.json`)

```json
{
  "i2c": { "bus": "2", "address": 72 },
  "sample_rate": 128,
  "outputs": [
    { "type": "console" },
    { "type": "mqtt", "mqtt": { "server": "tcp://localhost:1883", "client_id": "ads1115", "state_topic": "sensors/machine_battery/voltage", "discovery_topic": "homeassistant/sensor/machine_battery/config", "discovery_name": "Li-Ion machine battery", "discovery_unique_id": "machine_battery_lion_01" } }
  ],
  "sensor_type": "real",
  "channels": [
    { "channel": 0, "enabled": true,  "calibration_scale": 1.0, "calibration_offset": 0.0 },
    { "channel": 1, "enabled": false, "calibration_scale": 1.0, "calibration_offset": 0.0 },
    { "channel": 2, "enabled": false, "calibration_scale": 1.0, "calibration_offset": 0.0 },
    { "channel": 3, "enabled": false, "calibration_scale": 1.0, "calibration_offset": 0.0 }
  ]
}
```

### Fields

The table below is the authoritative reference for configuration fields and corresponding CLI flags. Command-line flags override values in the JSON file.

### Fields (table)

| Config (JSON path) | Flag | Description |
|---|---|---|
| `i2c.bus` | `-i2c-bus` | I2C bus to use (string). Example: `"2"` → `/dev/i2c-2`. Default: `"2"`. |
| `i2c.address` | `-i2c-address` | ADS1115 I2C address (decimal or `0x` hex). Default: `0x48` (72). |
| `sample_rate` | `-sample-rate` | Global ADS1115 conversion rate in SPS used as a default when a channel doesn't override it. Supported values: `8,16,32,64,128,250,475,860`. Default: `128`. |
| `channels[]` | `-channel-enabled` | Array of per-channel objects. Use `-channel-enabled` mappings to enable/disable channels (example: `-channel-enabled 0=true,1=false`). See per-field flags below. |
| `channels[].channel` | (none) | Channel index (0..3). |
| `channels[].enabled` | `-channel-enabled` | Whether this channel is read. Default: `false`. Accepts mappings like `0=true,1=false`. |
| `channels[].sample_rate` | `-channel-sample-rates` | Optional per-channel sample rate (SPS). Mapping example: `0=250,1=128`. If omitted, root `sample_rate` is used. |
| `channels[].calibration_scale` | `-channel-scales` | Per-channel multiplicative calibration factor. Mapping example: `0=1.0,1=0.98`. Default per-channel: `1.0`. |
| `channels[].calibration_offset` | `-channel-offsets` | Per-channel additive offset applied after scaling. Mapping example: `0=0.12,1=-0.05`. Default per-channel: `0.0`. |
| `outputs[].type` | `-outputs` | Output type: `console` or `mqtt`. CLI accepts CSV (e.g. `console,mqtt`) for quick config which creates basic entries. |
| `outputs[].interval_ms` | `-output-intervals` | Publish interval (ms) for this output. If omitted, a recommended interval is derived from enabled channels and their sample rates (approx: sum over enabled channels of `1000/sample_rate + 2ms`). Use `-output-intervals` CSV to set per-output values, e.g. `console=1000,mqtt=5000`. |
| `outputs[].mqtt.server` | `-mqtt-server` | MQTT broker URL (e.g. `tcp://host:1883`). Applied to all `mqtt` outputs; if none exist and flags provided, a `mqtt` output will be created. |
| `outputs[].mqtt.username` | `-mqtt-user` | MQTT username (optional). |
| `outputs[].mqtt.password` | `-mqtt-pass` | MQTT password (optional). |
| `outputs[].mqtt.client_id` | `-mqtt-client-id` | MQTT client id (optional). |
| `outputs[].mqtt.state_topic` | `-mqtt-state-topic` | State topic to publish readings under (e.g. `sensors/machine_battery/voltage`). When using Home Assistant MQTT discovery this value will be used as the `state_topic` in the discovery payload. |
| `outputs[].mqtt.discovery_topic` | `-mqtt-discovery-topic` | Full MQTT topic where Home Assistant discovery payload will be published (e.g. `homeassistant/sensor/machine_battery/config`). If empty, discovery is not published. |
| `sensor_type` | `-sensor-type` | `real` (ADS1115 via I2C) or `simulation` (fake sensor). Default: `real`. |
| `config` | `-config` | Path to JSON config file. Default: `./config.json` if present. Flags override file values. |

## Best practices

- For multiple channels, ensure `outputs[].interval_ms` is >= sensor read interval (derived from `sample_rate`) to avoid publishing identical snapshots repeatedly.
- `128` SPS is a good default for most use cases.

## Wiring (example)

Wiring the battery to the analog sensor (ADS1115)

```
    +-------------------+
    |   Li-ion battery  |
    |      +4.2V max    |
    +-------------------+
           |
           R1 = 10kΩ
           |
           +----> AINx (ADS1115)
           |
           R2 = 33kΩ
           |
          GND
```

This section shows a simple voltage divider example to scale battery voltage into an ADS1115 input. Adjust resistor values and ADC PGA as needed for your voltage range.

## Home Assistant integration

The MQTT output can publish a Home Assistant discovery configuration so the sensor is automatically discovered. The discovery message is published to the topic `homeassistant/sensor/<<object_id>>/config` and should contain a JSON payload like:

```json
{
  "name": "<<name>>",
  "state_topic": "<<stateTopic>>",
  "unit_of_measurement": "V",
  "device_class": "voltage",
  "state_class": "measurement",
  "value_template": "{{ value_json.voltage }}",
  "json_attributes_topic": "<<stateTopic>>",
}
```

Value messages are published to the `state_topic` (for example `sensors/machine_battery/voltage`) and should be JSON objects containing the reading. Example payload:

```json
{ "voltage": 3.72 }
```

Notes:
- The topic used for state updates can be configured via `outputs[].mqtt.state_topic` (CLI flag `-mqtt-state-topic`).
- When discovery is enabled the application will publish the discovery config with the `state_topic` so Home Assistant can read values from that topic.
 - The discovery topic (where the discovery JSON is published) can be configured via `outputs[].mqtt.discovery_topic` or CLI flag `-mqtt-discovery-topic`.
 - The topic used for state updates can be configured via `outputs[].mqtt.state_topic` (CLI flag `-mqtt-state-topic`).
 - When discovery is enabled the application will publish the discovery config with the `state_topic` so Home Assistant can read values from that topic.
 - The discovery topic (where the discovery JSON is published) can be configured via `outputs[].mqtt.discovery_topic` or CLI flag `-mqtt-discovery-topic`.
 - Note: `unit_of_measurement` and `device_class` in the discovery payload are fixed to `"V"` and `"voltage"` respectively and are not configurable.

## Contributing

This repository is a minimal starter. Please open issues or PRs to suggest improvements, add outputs, or fix bugs.
