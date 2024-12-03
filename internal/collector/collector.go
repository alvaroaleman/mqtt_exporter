package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

func New() *Collector {
	return &Collector{
		data: map[metricName]map[string]map[string]float64{
			metricNameHumidity:    {},
			metricNameTemperature: {},
			metricNameBattery:     {},
			metricNameLinkQuality: {},
			metricNameFertility:   {},
			metricNameLight:       {},
			metricNameMoisture:    {},
		},
	}
}

type metricName string

const (
	metricPrefix          metricName = "mqtt_"
	metricNameHumidity    metricName = metricPrefix + "humidity_percent"
	metricNameTemperature metricName = metricPrefix + "temperature_celsius"
	metricNameBattery     metricName = metricPrefix + "battery_percent"
	metricNameLinkQuality metricName = metricPrefix + "linkquality"
	metricNameFertility   metricName = metricPrefix + "fertility"
	metricNameLight       metricName = metricPrefix + "light_lux"
	metricNameMoisture    metricName = metricPrefix + "moisture_percent"
)

type Collector struct {
	// data maps metricName -> deviceType -> deviceName -> value
	data     map[metricName]map[string]map[string]float64
	dataLock sync.Mutex
}

func (c *Collector) SetHumidity(deviceType string, deviceName string, value float64) {
	c.set(metricNameHumidity, deviceType, deviceName, value)
}

func (c *Collector) SetTemperature(deviceType string, deviceName string, value float64) {
	c.set(metricNameTemperature, deviceType, deviceName, value)
}

func (c *Collector) SetBattery(deviceType string, deviceName string, value float64) {
	c.set(metricNameBattery, deviceType, deviceName, value)
}

func (c *Collector) SetLinkQuality(deviceType string, deviceName string, value float64) {
	c.set(metricNameLinkQuality, deviceType, deviceName, value)
}

func (c *Collector) SetFertility(deviceType string, deviceName string, value float64) {
	c.set(metricNameFertility, deviceType, deviceName, value)
}

func (c *Collector) SetLight(deviceType string, deviceName string, value float64) {
	c.set(metricNameLight, deviceType, deviceName, value)
}

func (c *Collector) SetMoisture(deviceType string, deviceName string, value float64) {
	c.set(metricNameMoisture, deviceType, deviceName, value)
}

func (c *Collector) set(name metricName, deviceType string, deviceName string, value float64) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	if c.data[name] == nil {
		c.data[name] = make(map[string]map[string]float64, 1)
	}
	if c.data[name][deviceType] == nil {
		c.data[name][deviceType] = make(map[string]float64, 1)
	}
	c.data[name][deviceType][deviceName] = value
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	for metricName := range c.data {
		ch <- prometheus.NewDesc(string(metricName), string(metricName), []string{"type", "name"}, nil)
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	for metricName, deviceTypes := range c.data {
		for deviceType, devices := range deviceTypes {
			for deviceName, value := range devices {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(string(metricName), string(metricName), []string{"type", "name"}, nil),
					prometheus.GaugeValue, value, deviceType, deviceName,
				)
			}
		}
	}
}
