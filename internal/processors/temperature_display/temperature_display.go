package temperaturedisplay

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alvaroaleman/mqtt_exporter/internal/config"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func New(log *zap.Logger, config config.Config, registry prometheus.Registerer) (processors.Processor, error) {
	temperature := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "temperature_celsius",
		Help:        "Temperature in Celsius",
		ConstLabels: prometheus.Labels{"type": "temperature_display"},
	}, []string{"name"})
	if err := registry.Register(temperature); err != nil {
		return nil, fmt.Errorf("failed to register temperature metric: %w", err)
	}
	humidity := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "humidity_percent",
		Help:        "Humidity",
		ConstLabels: prometheus.Labels{"type": "temperature_display"},
	}, []string{"name"})
	if err := registry.Register(humidity); err != nil {
		return nil, fmt.Errorf("failed to register moisture metric: %w", err)
	}
	battery := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "battery_percent",
		ConstLabels: prometheus.Labels{"type": "temperature_display"},
	}, []string{"name"})
	if err := registry.Register(battery); err != nil {
		return nil, fmt.Errorf("failed to register battery metric: %w", err)
	}
	linkquality := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "linkquality",
		ConstLabels: prometheus.Labels{"type": "temperature_display"},
	}, []string{"name"})
	if err := registry.Register(linkquality); err != nil {
		return nil, fmt.Errorf("failed to register linkquality metric: %w", err)
	}

	return &temperatureDisplayProcessor{
		log:         log,
		temperature: temperature,
		humidity:    humidity,
		battery:     battery,
		linkquality: linkquality,
	}, nil
}

type temperatureDisplayProcessor struct {
	log         *zap.Logger
	temperature *prometheus.GaugeVec
	humidity    *prometheus.GaugeVec
	battery     *prometheus.GaugeVec
	linkquality *prometheus.GaugeVec
}

func (t *temperatureDisplayProcessor) Describe(ch chan<- *prometheus.Desc) {
	t.temperature.Describe(ch)
	t.humidity.Describe(ch)
	t.battery.Describe(ch)
	t.linkquality.Describe(ch)
}

func (t *temperatureDisplayProcessor) Name() string {
	return "temperature_display"
}

func (t *temperatureDisplayProcessor) Process(topic string, msg []byte) bool {
	if !strings.HasPrefix(topic, "zigbee2mqtt/") {
		return false
	}
	sensorName := strings.Split(topic, "/")[1]
	var m temperatureDisplayMessage
	if err := json.Unmarshal(msg, &m); err != nil {
		t.log.Error("failed to unmarshal message", zap.Error(err))
		return true
	}
	t.temperature.WithLabelValues(sensorName).Set(m.Temperature)
	t.humidity.WithLabelValues(sensorName).Set(m.Humidity)
	t.battery.WithLabelValues(sensorName).Set(m.Battery)
	t.linkquality.WithLabelValues(sensorName).Set(m.Linkquality)

	t.log.Debug("Processed message", zap.String("topic", topic), zap.Any("message", m))
	return true
}

type temperatureDisplayMessage struct {
	Battery     float64 `json:"battery"`
	Humidity    float64 `json:"humidity"`
	Linkquality float64 `json:"linkquality"`
	Temperature float64 `json:"temperature"`
}
