// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

// Config -> config structure
type Config struct {
	ConfigDir string `yaml:"config_dir"`
	Period    string `yaml:"period"`
}

var (
	// DefaultConfig -> default values
	DefaultConfig = Config{
		ConfigDir: "jolokiabeat.d",
		Period:    "10s",
	}
)
