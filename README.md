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


## Configuration

### Example config (`config.json`)

```json
{
  "i2c_bus": "2",
  "i2c_address": 72,
  "sample_rate": 128,
  "calibration_scale": 1.0,
  "calibration_offset": 0.0,
  "outputs": [
    { "type": "console" },
    { "type": "mqtt", "mqtt": { "server": "tcp://localhost:1883", "client_id": "ads1115", "topic": "ads1115" } }
  ],
  "sensor_type": "real",
  "channels": [0,1,2,3]
}
```

### Fields

The table below is the authoritative reference for configuration fields and corresponding CLI flags. Command-line flags override values in the JSON file.

### Fields (table)

| Config (JSON path) | Flag | Description |
|---|---|---|
| `i2c_bus` | `-i2c-bus` | I2C bus to use (string). Example: `"2"` → `/dev/i2c-2`. Default: `"2"`. |
| `i2c_address` | `-i2c-address` | ADS1115 I2C address (decimal or `0x` hex). Default: `0x48` (72). |
| `sample_rate` | `-sample-rate` | ADS1115 conversion rate in SPS. Supported values: `8,16,32,64,128,250,475,860`. Default: `128`. Used to compute conversion time (~1000/sample_rate ms) and a recommended sensor read interval (≈ channels*(1000/sample_rate + 2ms)). |
| `calibration_scale` | `-calibration` | Multiplicative calibration factor applied to converted readings. Default: `1.0`. |
| `calibration_offset` | `-calibration-offset` | Additive calibration offset applied after scaling. Default: `0.0`. |
| `outputs[].type` | `-outputs` | Output type: `console` or `mqtt`. CLI accepts CSV (e.g. `console,mqtt`) for quick config which creates basic entries. |
| `outputs[].interval_ms` | `-output-intervals` | Publish interval (ms) for this output. If omitted, a recommended interval is derived from `sample_rate` and channels (approx: channels*(1000/sample_rate + 2ms)). Use `-output-intervals` CSV to set per-output values, e.g. `console=1000,mqtt=5000`. |
| `outputs[].mqtt.server` | `-mqtt-server` | MQTT broker URL (e.g. `tcp://host:1883`). Applied to all `mqtt` outputs; if none exist and flags provided, a `mqtt` output will be created. |
| `outputs[].mqtt.username` | `-mqtt-user` | MQTT username (optional). |
| `outputs[].mqtt.password` | `-mqtt-pass` | MQTT password (optional). |
| `outputs[].mqtt.client_id` | `-mqtt-client-id` | MQTT client id (optional). |
| `outputs[].mqtt.topic` | `-mqtt-topic` | Base topic to publish readings under (e.g. `ads1115`). |
| `sensor_type` | `-sensor-type` | `real` (ADS1115 via I2C) or `simulation` (fake sensor). Default: `real`. |
| `channels` | `-channels` | CSV or array of channel indices to read (0..3). Default: `0,1,2,3`. |
| `config` | `-config` | Path to JSON config file. Default: `./config.json` if present. Flags override file values. |

## Best practices

- For multiple channels, ensure `outputs[].interval_ms` is >= sensor read interval (derived from `sample_rate`) to avoid publishing identical snapshots repeatedly.
- `128` SPS is a good default for most use cases.

## Wiring

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

## Contributing

This repository is a minimal starter. Please open issues or PRs to suggest improvements, add outputs, or fix bugs.
