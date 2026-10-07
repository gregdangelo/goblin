package config

/*
Todo: fix naming in here
*/
type DelayMode string

const (
	DelayModeFixed  DelayMode = "fixed" // the only one I've actually implemented
	DelayModeRandom DelayMode = "random"
	DelayModeJitter DelayMode = "jitter" // small changes
)

// maybe reponse and transport re-use this but separate in Chaos Config?
type DelayConfigSettings struct {
	Enabled       bool      `toml:"enabled"`
	Mode          DelayMode `toml:"mode"`
	Min           int64     `toml:"min"`
	Max           int64     `toml:"max"`
	ResponseDelay int64     `toml:"response_delay"` //place holder
}

type DelayConfig struct {
	Transport DelayConfigSettings `toml:"transport"`
	Response  DelayConfigSettings `toml:"response"`
}

type FailRequestsConfig struct {
	Enabled bool `toml:"enabled"`
}

type ChaosConfig struct {
	DelayConfig        `toml:"delay_config"`
	FailRequestsConfig `toml:"fail_requests"`
}
