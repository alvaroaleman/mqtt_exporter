package processors

type Processor interface {
	Process(topic string, msg []byte) (handled bool)
	Name() string
}
