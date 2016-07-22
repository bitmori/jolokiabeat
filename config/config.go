package config

// Config -> config structure
type Config struct {
	Jolokiabeat JolokiabeatConfig
}

// JolokiabeatConfig -> beat config
type JolokiabeatConfig struct {
	ConfigDir       string `config:"config_dir"`
	Period          string `config:"period"`
	FieldName       string `config:"metric_field_name"`
	MetricUnderRoot bool   `config:"metric_under_root"`
}

var (
	// DefaultConfig -> default values
	DefaultConfig = Config{
		Jolokiabeat: JolokiabeatConfig{
			ConfigDir:       "jolokiabeat.d",
			Period:          "10s",
			FieldName:       "bean",
			MetricUnderRoot: false,
		},
	}
)

// CheckConfig -> make sure the default value is loaded
func (config *Config) CheckConfig() {
	if config.Jolokiabeat.ConfigDir == "" {
		config.Jolokiabeat.ConfigDir = DefaultConfig.Jolokiabeat.ConfigDir
	}

	if config.Jolokiabeat.FieldName == "" {
		config.Jolokiabeat.FieldName = DefaultConfig.Jolokiabeat.FieldName
	}

	if config.Jolokiabeat.Period == "" {
		config.Jolokiabeat.Period = DefaultConfig.Jolokiabeat.Period

	}
}
