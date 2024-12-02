package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alvaroaleman/mqtt_exporter/internal/config"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors"
	"github.com/alvaroaleman/mqtt_exporter/internal/processors/miflora"
	temperaturedisplay "github.com/alvaroaleman/mqtt_exporter/internal/processors/temperature_display"
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
	}

	gatherers := prometheus.Gatherers{prometheus.DefaultGatherer}

	var processors []processors.Processor
	miflora, err := miflora.New(log, config, prometheus.DefaultRegisterer)
	if err != nil {
		return fmt.Errorf("failed to construct miflora processor: %w", err)
	}
	processors = append(processors, miflora)

	temperaturedisplayRegistry := prometheus.NewRegistry()
	temperatureDisplayProcessor, err := temperaturedisplay.New(log, config, temperaturedisplayRegistry)
	if err != nil {
		return fmt.Errorf("failed to construct temperature display processor: %w", err)
	}
	gatherers = append(gatherers, temperaturedisplayRegistry)
	processors = append(processors, temperatureDisplayProcessor)

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
					log.Debug("Processed message", zap.String("processor", processor.Name()))
					break
				}
			}
			log.Debug("Received message",
				zap.String("topic", m.Topic()),
			//	zap.ByteString("payload", m.Payload()),
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

	http.Handle("/metrics", promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{}),
	))

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
