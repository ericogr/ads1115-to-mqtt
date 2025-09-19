# ADS1115 to MQTT

Leitura de um ADS1115 (4 canais) e envio para saídas configuráveis (console, MQTT).

Principais recursos:
- Configuração via arquivo JSON (`-config path`) ou via flags de linha de comando
- Suporta sensor real (ADS1115 via I2C) ou sensor de simulação
- Saídas: `console` e `mqtt` (é fácil adicionar outras)

Exemplos:

- Usar flags:
  - `./ads1115-to-mqtt -i2c-bus 2 -i2c-address 0x48 -outputs console,mqtt -mqtt-server tcp://mqtt.local:1883`
  - Exemplo com intervalos por saída: `./ads1115-to-mqtt -outputs console,mqtt -output-intervals console=1000,mqtt=5000 -mqtt-server tcp://broker:1883`
- Usar arquivo JSON:
  - `./ads1115-to-mqtt -config config.json`

Build:
- `make build` (nativo)
- `make build-dietpi` (Linux ARM64 para Orange Pi)
- `make build-linux` (Linux amd64)
- `make build-windows` (Windows amd64)

Run (development)

- `make run` — executa o programa com `-outputs console -sensor-type simulation` por padrão (útil para desenvolvimento/simulação).
  - Para personalizar: `make run RUN_ARGS="-outputs mqtt -mqtt-server tcp://broker:1883 -sensor-type simulation"`

Config (exemplo `config.json`):

```
{
  "i2c_bus": "2",
  "i2c_address": 72,
  "sample_rate": 128,
  "calibration_scale": 1.0,
  "calibration_offset": 0.0,
  "outputs": [
    { "type": "console", "interval_ms": 1000 },
    { "type": "mqtt", "interval_ms": 1000, "mqtt": { "server": "tcp://localhost:1883", "client_id": "ads1115", "topic": "ads1115" } }
  ],
  "sensor_type": "real",
  "channels": [0,1,2,3],
  "interval_ms": 1000
}
```

Configuration parameters

| Parâmetro (config.json) | Flag (linha de comando) | Descrição |
|---|---|---|
| `i2c_bus` | `-i2c-bus` | Barramento I2C (ex.: `2` -> `/dev/i2c-2`). |
| `i2c_address` | `-i2c-address` | Endereço I2C (decimal ou hex, ex.: `0x48`). |
| `sample_rate` | `-sample-rate` | Taxa de amostragem do ADS1115 em SPS (ex.: `128`). |
| `calibration_scale` | `-calibration` | Fator multiplicativo de calibração aplicado ao valor lido. |
| `calibration_offset` | `-calibration-offset` | Offset aditivo aplicado após a calibração. |
| `outputs[].type` | `-outputs` | Tipo da saída: `console` ou `mqtt`. Pode ser uma lista (ex.: `console,mqtt`). |
| `outputs[].interval_ms` | `-output-intervals` | Intervalo de publicação para esta saída (ms). Quando não informado, usa `interval_ms` global. Use flags como `console=1000,mqtt=5000`.
| `outputs[].mqtt.server` | `-mqtt-server` | Endereço do broker MQTT (ex.: `tcp://host:1883`). Aplicado a todos os outputs do tipo `mqtt`; se não houver, um `mqtt` output será criado. |
| `outputs[].mqtt.username` | `-mqtt-user` | Usuário MQTT (opcional). |
| `outputs[].mqtt.password` | `-mqtt-pass` | Senha MQTT (opcional). |
| `outputs[].mqtt.client_id` | `-mqtt-client-id` | Client ID usado pelo cliente MQTT. |
| `outputs[].mqtt.topic` | `-mqtt-topic` | Tópico base para publicação (ex.: `ads1115`). |
| `sensor_type` | `-sensor-type` | Tipo de sensor: `real` ou `simulation` (fake). |
| `channels` | `-channels` | Lista de canais a ler (ex.: `0,1,2,3`). |
| `interval_ms` | `-interval-ms` | Intervalo entre leituras do sensor em milissegundos. |
| (top-level) `config` | `-config` | Caminho para arquivo JSON de configuração que sobrescreve valores padrão. |
