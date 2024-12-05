package zigbee2mqtt

import (
	"encoding/json"
	"strings"

	"go.uber.org/zap"

	"github.com/alvaroaleman/mqtt_exporter/internal/collector"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors"
)

func New(log *zap.Logger, collector *collector.Collector) processors.Processor {
	return &temperatureDisplayProcessor{
		log:       log,
		collector: collector,
	}
}

type temperatureDisplayProcessor struct {
	log       *zap.Logger
	collector *collector.Collector
}

func (t *temperatureDisplayProcessor) Name() string {
	return "zigbee2mqtt"
}

func (t *temperatureDisplayProcessor) Process(topic string, msg []byte) bool {
	if !strings.HasPrefix(topic, "zigbee2mqtt/") || strings.HasPrefix(topic, "zigbee2mqtt/bridge") {
		return false
	}
	sensorName := strings.Split(topic, "/")[1]
	var m message
	if err := json.Unmarshal(msg, &m); err != nil {
		t.log.Error("failed to unmarshal message", zap.Error(err))
		return true
	}

	if m.Battery != nil {
		t.collector.SetBattery("sensor", sensorName, *m.Battery)
	}

	if m.Humidity != nil {
		t.collector.SetHumidity("sensor", sensorName, *m.Humidity)
	}

	if m.Linkquality != nil {
		t.collector.SetLinkQuality("sensor", sensorName, *m.Linkquality)

	}

	if m.Temperature != nil {
		t.collector.SetTemperature("sensor", sensorName, *m.Temperature)
	}

	t.log.Debug("Processed message", zap.String("topic", topic), zap.Any("message", m))
	return true
}

type message struct {
	Battery     *float64 `json:"battery"`
	Humidity    *float64 `json:"humidity"`
	Linkquality *float64 `json:"linkquality"`
	Temperature *float64 `json:"temperature"`
}
