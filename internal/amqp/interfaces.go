package amqp

type incomingMessage interface {
	GetMessage() []byte
	Successfully()
	Unsuccessfully()
}
