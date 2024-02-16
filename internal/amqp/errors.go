package amqp

import (
	"fmt"
)

type envStreamIError struct {
}

func (e envStreamIError) Error() string {
	return "envStream error"
}

type envStreamIsClosedError struct {
}

func (e envStreamIsClosedError) Error() string {
	return "envStream is closed"
}

type envStreamNotDefineError struct {
}

func (e envStreamNotDefineError) Error() string {
	return "envStream is not define"
}

type createConnRabbitError struct {
	errMessage string
}

func (e createConnRabbitError) Error() string {
	return "error create connect RabbitMQ"
}

type createProducerError struct {
}

func (e createProducerError) Error() string {
	return "error create producer RabbitMQ"
}

type producerNotDefineError struct {
}

func (e producerNotDefineError) Error() string {
	return "producerStream is not define"
}

type closeProducerError struct {
}

func (e closeProducerError) Error() string {
	return "error close producer rabbit"
}

type closeConnError struct {
}

func (e closeConnError) Error() string {
	return "error close conn rabbit"
}

type messageConversionError struct {
	place string
}

func (e messageConversionError) Error() string {
	return fmt.Sprintf("message conversion error: %s", e.place)
}

type messageSendError struct {
}

func (e messageSendError) Error() string {
	return "message could not be sent"
}
