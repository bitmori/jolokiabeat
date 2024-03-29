package beater

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	jConfig "github.com/neonmori/jolokiabeat/config"
)

const selector = "Jolokiabeat"

// Jolokiabeat collects JMX info reported by jolokia
type Jolokiabeat struct {
	done            chan struct{}
	config          jConfig.Config
	jolokia         JolokiaToml
	publisher       publisher.Client
	requester       *http.Client
	period          time.Duration
	metricUnderRoot bool
	metricFieldName string
}

// New -> Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := jConfig.Config{}
	if err := cfgfile.Read(&config, ""); err != nil {
		return nil, err
	}
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	config.CheckConfig()

	logp.Debug(selector, "%v", config)
	var jolokia JolokiaToml
	if err := jolokia.LoadDirectory(config.Jolokiabeat.ConfigDir); err != nil {
		return nil, fmt.Errorf("Error loading jolokia configs: %v", err)
	}

	ego := &Jolokiabeat{
		metricFieldName: config.Jolokiabeat.FieldName,
		metricUnderRoot: config.Jolokiabeat.MetricUnderRoot,
		done:            make(chan struct{}),
		config:          config,
	}
	ego.jolokia = jolokia
	var err error
	ego.period, err = time.ParseDuration(config.Jolokiabeat.Period)
	if err != nil {
		return nil, err
	}
	_tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
	ego.requester = &http.Client{
		Transport: _tr,
		Timeout:   time.Duration(4 * time.Second),
	}
	ego.publisher = b.Publisher.Connect()

	return ego, nil
}

// collect goroutine
func (ego *Jolokiabeat) collect(j Jok, s Server) {
	ticker := time.NewTicker(ego.period)
	defer ticker.Stop()
	for {
		select {
		case <-ego.done:
			return
		case <-ticker.C:
		}
		if err := ego.CollectData(j, s); err != nil {
			logp.Err("Error while getting JMX: %v", err)
		}
	}
}

// Run -> Main goroutine of jolokiabeat
func (ego *Jolokiabeat) Run(b *beat.Beat) error {
	logp.Info("jolokiabeat is running! Hit CTRL-C to stop it.")

	// for each jok ->
	for _, j := range ego.jolokia.Jolokia {
		// for each server ->
		for _, server := range j.Servers {
			go ego.collect(j, server)
		}
	}

	<-ego.done
	return nil
}

// Stop -> callback when jolokiabeat stops
func (ego *Jolokiabeat) Stop() {
	ego.publisher.Close()
	close(ego.done)
}
