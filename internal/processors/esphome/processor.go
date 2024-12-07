package esphome

import (
	"strconv"
	"strings"

	"github.com/alvaroaleman/mqtt_exporter/internal/collector"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors"
	"go.uber.org/zap"
)

func New(log *zap.Logger, collector *collector.Collector) processors.Processor {
	return &esphomeProcessor{
		log:       log,
		collector: collector,
	}
}

type esphomeProcessor struct {
	log       *zap.Logger
	collector *collector.Collector
}

func (m *esphomeProcessor) Name() string {
	return "esphome"
}

func (m *esphomeProcessor) Process(topic string, msg []byte) (handled bool) {
	if !strings.Contains(topic, "/sensor/") || !strings.HasSuffix(topic, "/state") {
		return false
	}

	val, err := strconv.ParseFloat(string(msg), 64)
	if err != nil {
		m.log.Error("failed to parse message as float", zap.Error(err), zap.ByteString("raw", msg))
		return

	}
	topicNameComponents := strings.Split(topic, "/")
	sensorName := topicNameComponents[(len(topicNameComponents) - 2)]

	switch {
	case strings.HasSuffix(sensorName, "_temperature"):
		m.collector.SetTemperature("plant", strings.TrimSuffix(sensorName, "_temperature"), val)
	case strings.HasSuffix(sensorName, "_soil_conductivity"):
		m.collector.SetFertility("plant", strings.TrimSuffix(sensorName, "_soil_conductivity"), val)
	case strings.HasSuffix(sensorName, "_illuminance"):
		m.collector.SetLight("plant", strings.TrimSuffix(sensorName, "_illuminance"), val)
	case strings.HasSuffix(sensorName, "_moisture"):
		m.collector.SetMoisture("plant", strings.TrimSuffix(sensorName, "_moisture"), val)
	default:
		m.log.Debug("unknown sensor type", zap.String("sensor", sensorName))
	}

	return true
}
