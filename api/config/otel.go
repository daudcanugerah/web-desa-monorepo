package config

type OtelConfig struct {
	EnableMetric bool `mapstructure:"metric"`
	EnableLog    bool `mapstructure:"log"`
	EnableTrace  bool `mapstructure:"trace"`
}
