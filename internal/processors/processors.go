package processors

type Processor interface {
	Process(msg []byte) (handled bool)
}
