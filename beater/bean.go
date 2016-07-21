package beater

import (
	"io/ioutil"
	"path/filepath"

	"github.com/naoina/toml"
)

// Server -> struct of jolokia server
type Server struct {
	Name     string
	Host     string
	Username string
	Password string
	Port     string
}

// Metric -> struct of jmx metric
type Metric struct {
	Name      string
	Mbean     string
	Attribute string
	Path      string
	Jmx       string // this won't be supported if we use POST requests
}

// Jok -> A single jolokia config set
type Jok struct {
	Context string
	Mode    string
	Servers []Server
	Metrics []Metric
	Proxy   Server
}

// JolokiaToml -> Toml root
type JolokiaToml struct {
	Jolokia []Jok
}

// LoadDirectory -> load config in toml files in jolokiabeat.d folder
func (ego *JolokiaToml) LoadDirectory(path string) error {
	directoryEntries, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 6 || name[len(name)-5:] != ".toml" {
			continue
		}
		err := ego.loadConfig(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (ego *JolokiaToml) loadConfig(path string) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var config JolokiaToml
	err = toml.Unmarshal(contents, &config)
	if err != nil {
		return err
	}
	ego.Jolokia = append(ego.Jolokia, config.Jolokia...)
	return nil
}
