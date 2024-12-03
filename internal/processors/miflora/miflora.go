package miflora

import (
	"encoding/json"

	"go.uber.org/zap"

	"github.com/alvaroaleman/mqtt_exporter/internal/collector"
	"github.com/alvaroaleman/mqtt_exporter/internal/config"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors"
)

func New(log *zap.Logger, config config.Config, collector *collector.Collector) (processors.Processor, error) {
	return &miFloraProcessor{
		log:       log,
		config:    config,
		collector: collector,
	}, nil
}

type miFloraProcessor struct {
	log       *zap.Logger
	config    config.Config
	collector *collector.Collector
}

func (m *miFloraProcessor) Name() string {
	return "miflora"
}

func (m *miFloraProcessor) Process(_ string, msg []byte) (handled bool) {
	var target miFloraMessage
	if err := json.Unmarshal(msg, &target); err != nil {
		m.log.Debug("failed to unmarshal message as miFloraMessage", zap.Error(err), zap.ByteString("raw", msg))
		return false
	}

	if target.ID == "" {
		return false
	}

	name := target.ID
	if m.config.Configs[name].HumanReadableName != "" {
		name = m.config.Configs[name].HumanReadableName
	}

	if target.TempCelius != nil {
		m.collector.SetTemperature("plant", name, *target.TempCelius)
	}
	if target.Fertility != nil {
		m.collector.SetFertility("plant", name, *target.Fertility)
	}
	if target.Light != nil {
		m.collector.SetLight("plant", name, *target.Light)
	}
	if target.Moisture != nil {
		m.collector.SetMoisture("plant", name, *target.Moisture)
	}

	return true
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
