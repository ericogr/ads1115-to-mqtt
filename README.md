# ADS1115 to MQTT

Leitura analógica via ADS1115 (4 canais) e publicação das leituras em saídas configuráveis (console, MQTT).

Visão geral
- Leitura do ADS1115 por I2C (ou sensor de simulação para testes).
- Saídas configuráveis por item (`outputs[]`) — cada saída tem `type` e `interval_ms`.
- Configuração por arquivo JSON (`./config.json` por padrão) e/ou flags; flags têm precedência.

Principais usos
- Testes/Desenvolvimento: `make run` (usa sensor fake + saída `console`).
- Produção: `make build` e executar o binário com `-config` ou flags.

Quickstart

1. Executar em modo dev (console + simulação):

   make run

2. Compilar:

   make build

3. Executar binário (exemplo):

   ./bin/ads1115-to-mqtt -outputs console,mqtt -output-intervals console=1000,mqtt=5000 -mqtt-server tcp://broker:1883

Arquivo de exemplo (`config.json`)

```
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

Esquema (config.json) — o que cada campo faz

- `i2c_bus` (string): barramento I2C a usar (ex.: `"2"` → `/dev/i2c-2`).
- `i2c_address` (int/hex): endereço I2C do ADS1115 (ex.: `72` ou `0x48`).
- `sample_rate` (int): taxa de conversão do ADS1115 em SPS. Valores suportados: `8, 16, 32, 64, 128, 250, 475, 860`. Default: `128`.
  - Serve para calcular o tempo de conversão: cada conversão ≈ `1000/sample_rate` ms.
  - O programa calcula automaticamente um intervalo de leitura recomendado (sensor interval) com base em `sample_rate` e no número de canais: aproximadamente `channels * (1000/sample_rate + 2ms)`.
- `calibration_scale` (float): multiplicador aplicado ao valor convertido (ajuste de ganho).
- `calibration_offset` (float): valor aditivo aplicado após o scale (ajuste de offset).
- `outputs` (array): lista de saídas; cada item é um objeto:
  - `type` (string): `console` ou `mqtt`.
  - `interval_ms` (int, opcional): intervalo de publicação para essa saída em ms. Se omitido, usa o intervalo recomendado calculado a partir de `sample_rate` e `channels`.
  - `mqtt` (obj, opcional): quando `type == "mqtt"`, contém `server`, `username`, `password`, `client_id`, `topic`.
- `sensor_type` (string): `real` (ADS1115 via I2C) ou `simulation` (sensor fake para testes).
- `channels` (array[int]): canais a ler (0..3).

Observações sobre intervals
- `sample_rate` define quão rápido o ADC converte; o programa agrupa leituras por chamada a `sensor.Read()` — que realiza uma conversão por canal sequencialmente.
- Cada saída publica o último snapshot de leituras no seu `outputs[].interval_ms`. Se uma saída publicar mais rápido que o sensor, ela reenviará o mesmo snapshot até haver novas leituras.

Flags (mapa rápido)

| Config JSON | Flag | Observação |
|---|---:|---|
| `i2c_bus` | `-i2c-bus` | Barramento I2C |
| `i2c_address` | `-i2c-address` | Endereço I2C (decimal ou hex)
| `sample_rate` | `-sample-rate` | SPS do ADS1115 (valores suportados acima)
| `calibration_scale` | `-calibration` | Scale multiplicativo
| `calibration_offset` | `-calibration-offset` | Offset aditivo
| `outputs[].type` | `-outputs` | CSV de tipos compatível (ex.: `console,mqtt`) — cria entradas básicas
| `outputs[].interval_ms` | `-output-intervals` | Map CSV `console=1000,mqtt=5000`
| `outputs[].mqtt.server` | `-mqtt-server` | Broker MQTT (aplicado a todos os mqtt outputs; cria um se necessário)
| `outputs[].mqtt.username` | `-mqtt-user` | Usuário MQTT
| `outputs[].mqtt.password` | `-mqtt-pass` | Senha MQTT
| `outputs[].mqtt.client_id` | `-mqtt-client-id` | Client ID MQTT
| `outputs[].mqtt.topic` | `-mqtt-topic` | Tópico base MQTT
| `sensor_type` | `-sensor-type` | `real` ou `simulation`
| `channels` | `-channels` | Lista de canais (ex.: `0,1,2,3`)
| (arquivo) `config` | `-config` | Caminho para arquivo JSON (default: `./config.json` se existir)

Boas práticas / recomendações
- Para múltiplos canais, escolha `sample_rate` e `outputs[].interval_ms` de forma consistente: garantir `outputs[].interval_ms >= sensor_interval` evita republicações do mesmo snapshot.
- `128` SPS costuma ser um bom ponto de partida.

Suporte e contribuição
- Projeto minimalista: abra issues/PRs com sugestões de melhorias, novos outputs ou correções.

