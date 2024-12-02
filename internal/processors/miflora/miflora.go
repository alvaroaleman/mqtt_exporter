package miflora

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/alvaroaleman/mqtt_exporter/internal/config"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors"
)

func New(log *zap.Logger, registry prometheus.Registerer) (processors.Processor, error) {
	temperature := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "temperature_celsius",
		Help:        "Temperature in Celsius",
		ConstLabels: prometheus.Labels{"type": "plant"},
	}, []string{"name"})
	if err := registry.Register(temperature); err != nil {
		return nil, fmt.Errorf("failed to register temperature metric: %w", err)
	}
	fertility := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "fertility",
		Help:        "Fertility",
		ConstLabels: prometheus.Labels{"type": "plant"},
	}, []string{"name"})
	if err := registry.Register(fertility); err != nil {
		return nil, fmt.Errorf("failed to register fertility metric: %w", err)
	}
	light := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "light_lux",
		Help:        "Light",
		ConstLabels: prometheus.Labels{"type": "plant"},
	}, []string{"name"})
	if err := registry.Register(light); err != nil {
		return nil, fmt.Errorf("failed to register light metric: %w", err)
	}
	moisture := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "moisture_percent",
		Help:        "Moisture",
		ConstLabels: prometheus.Labels{"type": "plant"},
	}, []string{"name"})
	if err := registry.Register(moisture); err != nil {
		return nil, fmt.Errorf("failed to register moisture metric: %w", err)
	}
	return &miFloraProcessor{
		log:         log,
		state:       make(map[string]*miFloraState),
		temperature: temperature,
		fertility:   fertility,
		light:       light,
		moisture:    moisture,
	}, nil
}

type miFloraProcessor struct {
	log         *zap.Logger
	config      config.Config
	state       map[string]*miFloraState
	temperature *prometheus.GaugeVec
	fertility   *prometheus.GaugeVec
	light       *prometheus.GaugeVec
	moisture    *prometheus.GaugeVec
}

func (m *miFloraProcessor) Process(msg []byte) (handled bool) {
	var target miFloraMessage
	if err := json.Unmarshal(msg, &target); err != nil {
		m.log.Debug("failed to unmarshal message as miFloraMessage", zap.Error(err), zap.ByteString("raw", msg))
		return false
	}

	name := target.ID
	if m.config.Configs[name].HumanReadableName != "" {
		name = m.config.Configs[name].HumanReadableName
	}

	if m.state[name] == nil {
		m.state[name] = &miFloraState{}
	}
	if target.TempCelius != nil {
		m.state[name].TempCelius = target.TempCelius
	}
	if target.Fertility != nil {
		m.state[name].Fertility = target.Fertility
	}
	if target.Light != nil {
		m.state[name].Light = target.Light
	}
	if target.Moisture != nil {
		m.state[name].Moisture = target.Moisture
	}
	m.collect()

	return true
}

func (m *miFloraProcessor) collect() {
	for name, state := range m.state {
		if state.TempCelius != nil {
			m.temperature.WithLabelValues(name).Set(*state.TempCelius)
		}
		if state.Fertility != nil {
			m.fertility.WithLabelValues(name).Set(*state.Fertility)
		}
		if state.Light != nil {
			m.light.WithLabelValues(name).Set(*state.Light)
		}
		if state.Moisture != nil {
			m.moisture.WithLabelValues(name).Set(*state.Moisture)
		}
	}
}

type miFloraState struct {
	Name           string
	TempCelius     *float64
	Tempfahrenheit *float64
	Fertility      *float64
	Light          *float64
	Moisture       *float64
}

type miFloraMessage struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Rssi       int      `json:"rssi"`
	Brand      string   `json:"brand"`
	Model      string   `json:"model"`
	ModelID    string   `json:"model_id"`
	Type       string   `json:"type"`
	TempCelius *float64 `json:"tempc"`
	Fertility  *float64 `json:"fer"`
	Light      *float64 `json:"lux"`
	Moisture   *float64 `json:"moi"`
	Mac        string   `json:"mac"`
}
