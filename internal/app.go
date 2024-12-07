package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alvaroaleman/mqtt_exporter/internal/collector"
	"github.com/alvaroaleman/mqtt_exporter/internal/config"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors/esphome"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors/miflora"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors/zigbee2mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

type Opts struct {
	ServerAddress string
	ServerPort    int
	Topics        []string
	ConfigFile    string
}

func Run(opts Opts, log *zap.Logger) error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var config config.Config
	if opts.ConfigFile != "" {
		content, err := os.ReadFile(opts.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to read config file %s: %w", opts.ConfigFile, err)
		}
		if err := yaml.Unmarshal(content, &config); err != nil {
			return fmt.Errorf("failed to unmarshal config file: %w", err)
		}
		log.Info("Loaded configuration from file", zap.String("file", opts.ConfigFile))
	}

	collector := collector.New()
	if err := prometheus.Register(collector); err != nil {
		return fmt.Errorf("failed to register collector: %w", err)
	}

	processors := []processors.Processor{
		zigbee2mqtt.New(log, collector),
		esphome.New(log, collector),
	}

	miflora, err := miflora.New(log, config, collector)
	if err != nil {
		return fmt.Errorf("failed to construct miflora processor: %w", err)
	}
	processors = append(processors, miflora)

	mqttOps := mqtt.
		NewClientOptions().
		SetClientID("mqtt_exporter").
		AddBroker(fmt.Sprintf("tcp://%s:%d", opts.ServerAddress, opts.ServerPort)).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Info("Connected to MQTT server")
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Error("Connection lost", zap.Error(err))
		}).
		SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
			for _, processor := range processors {
				if processor.Process(m.Topic(), m.Payload()) {
					log.Info("Processed message", zap.String("processor", processor.Name()))
					break
				}
			}
			log.Debug("Received message",
				zap.String("topic", m.Topic()),
				zap.ByteString("payload", m.Payload()),
			)
		})

	client := mqtt.NewClient(mqttOps)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT server: %w", token.Error())
	}

	for _, topic := range opts.Topics {
		if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
			return fmt.Errorf("failed to subscribe to topics: %w", token.Error())
		}
	}
	log.Info("Successfully subscribed to all topics")

	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{Addr: ":8080", Handler: http.DefaultServeMux}
	go func() {
		log.Info("Starting HTTP server", zap.Int("port", 8080))
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error", zap.Error(err))
			cancel()
			return
		}
		log.Info("HTTP server shut down")
	}()

	<-ctx.Done()
	log.Info("Signal received, shutting down")
	client.Disconnect(250)
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shut down HTTP server: %w", err)
	}

	return nil
}
