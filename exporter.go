package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	appConf *Config
)

type Config struct {
	LabelMetrics []map[string]string `yaml:"label_metrics"`
	MockMetrics  []MockMetric        `yaml:"mock_metrics"`
}

type LabelMetric struct {
	ResourceInfo map[string]string `yaml:"resource_info"`
}

type MockMetric struct {
	Name   string            `yaml:"name"`
	Type   string            `yaml:"type"`
	Value  int               `yaml:"value"`
	Labels map[string]string `yaml:"labels"`
}

func loadConfig(path string) (conf *Config, err error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(f), &conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func randomNumber() float64 {
	seed := rand.NewSource(time.Now().UnixNano())
	randWithSeed := rand.New(seed)
	return randWithSeed.Float64()
}

func extractLabelMetrics() {
	for _, lm := range appConf.LabelMetrics {
		promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "mock_tag_info",
			Help:        "mock_tag_info",
			ConstLabels: lm,
		}).Set(1)
	}
}

func newCounter(mock MockMetric) {
	metric := promauto.NewCounter(prometheus.CounterOpts{
		Name:        mock.Name,
		Help:        mock.Name,
		ConstLabels: mock.Labels,
	})
	metric.Add(randomNumber() * 10) // increase result by one decimal point
}

func newGauge(mock MockMetric) {
	metric := promauto.NewGauge(prometheus.GaugeOpts{
		Name:        mock.Name,
		Help:        mock.Name,
		ConstLabels: mock.Labels,
	})
	metric.Set(float64(mock.Value))
}

func extractMockMetrics() {
	for _, mm := range appConf.MockMetrics {
		switch mm.Type {
		case "counter":
			newCounter(mm)
		case "gauge":
			newGauge(mm)
		}
	}
}

func recordMetrics() {
	go func() {
		extractLabelMetrics()
		extractMockMetrics()
	}()
}

func main() {
	config, err := loadConfig("config.yml")
	if err != nil {
		log.Errorf("error loading config: %s", err.Error())
		os.Exit(1)
	}
	appConf = config

	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())

	log.Infof("prometheus-mock-exporter listening on port %v", 2112)
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
