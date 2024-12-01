package internal

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type Opts struct {
	ServerAddress string
	ServerPort    int
	Topics        []string
}

func Run(opts Opts, log *zap.Logger) error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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
			log.Info("Received message",
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

	<-ctx.Done()
	log.Info("Signal received, shutting down")
	client.Disconnect(250)

	return nil
}
