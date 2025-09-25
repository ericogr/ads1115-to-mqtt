package sensor

import "github.com/ericogr/ads1115-to-mqtt/pkg/config"

// buildChannelSettings extracts common per-channel settings from the config.
// Returned maps contain an entry for every configured channel.
func buildChannelSettings(cfg config.Config) (channels []int, scales map[int]float64, offsets map[int]float64, sampleRates map[int]int) {
	channels = make([]int, 0)
	scales = make(map[int]float64)
	offsets = make(map[int]float64)
	sampleRates = make(map[int]int)
	for _, c := range cfg.Channels {
		scales[c.Channel] = c.CalibrationScale
		offsets[c.Channel] = c.CalibrationOffset
		if c.SampleRate != 0 {
			sampleRates[c.Channel] = c.SampleRate
		}
		if c.Enabled {
			channels = append(channels, c.Channel)
		}
	}
	return
}
