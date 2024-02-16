package amqp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/message"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"

	log "github.com/sirupsen/logrus"
)

func InitRabbitWorker(url, nameProducer, nameStream string) *RabbitWorker {
	ctx, cancel := context.WithCancel(context.Background())

	r := &RabbitWorker{
		url:          url,
		nameProducer: nameProducer,
		nameStream:   nameStream,
		incomingCh:   make(chan interface{}),
		cancel:       cancel,
	}

	go func() {
		go r.createInterfaces()
		time.Sleep(5 * time.Second)
		go r.producerProcess(ctx)

	}()

	return r
}

type RabbitWorker struct {
	url            string
	nameProducer   string
	nameStream     string
	streamEnv      *stream.Environment
	streamProducer *stream.Producer
	incomingCh     chan interface{}
	cancel         func()
}

func (r *RabbitWorker) GetCh() chan interface{} {
	return r.incomingCh
}

func (r *RabbitWorker) createInterfaces() {
	err := r.createConn()
	if err != nil {
		return
	}

	err = r.createProducer()
	if err != nil {
		return
	}
}

// создание подключения к rabbit
func (r *RabbitWorker) createConn() error {
	streamOptions := stream.NewEnvironmentOptions().SetUri(r.url)
	env, err := stream.NewEnvironment(streamOptions)

	if err != nil {
		err = fmt.Errorf("%w: %w", createConnRabbitError{}, err)
		log.Error(err)
		return err
	}

	r.streamEnv = env
	log.Info("streamEnv created")
	return err
}

func (r *RabbitWorker) allCheck() error {
	err := r.checkConn()
	if err != nil {
		err = fmt.Errorf("%w: %w", envStreamIError{}, err)
		return err
	}

	err = r.checkProducer()
	if err != nil {
		return err
	}

	return err
}

func (r *RabbitWorker) checkConn() error {
	var err error

	if r.streamEnv == nil {
		err = envStreamNotDefineError{}
		log.Error(err)
		return err
	}

	if r.streamEnv.IsClosed() {
		err = envStreamIsClosedError{}
		log.Error(err)
		return err
	}

	return err
}

func (r *RabbitWorker) checkProducer() error {
	var err error
	if r.streamProducer == nil {
		err = producerNotDefineError{}
		log.Error(err)
	}

	return err
}

func (r *RabbitWorker) createProducer() error {
	producerOption := &stream.ProducerOptions{
		Name:                 r.nameProducer,
		QueueSize:            2000,
		BatchSize:            10000,
		BatchPublishingDelay: 500,
		SubEntrySize:         300,
	}

	producer, err := r.streamEnv.NewProducer(r.nameStream, producerOption)

	if err != nil {
		err = fmt.Errorf("%w: %w", createProducerError{}, err)
		log.Error(err)
		return err
	}

	r.streamProducer = producer
	log.Info("streamProducer created")
	return err
}

func (r *RabbitWorker) sendMessage(msg []byte) error {
	var message message.StreamMessage
	var err error

	err = r.allCheck()
	if err != nil {
		return err
	}

	message = amqp.NewMessage(msg)
	err = r.streamProducer.Send(message)
	return err
}

func (r *RabbitWorker) closeProducer() error {
	err := r.checkProducer()
	if err != nil {
		err = fmt.Errorf("%w: %w", closeProducerError{}, err)
		return err
	}

	err = r.streamProducer.Close()
	if err != nil {
		err = fmt.Errorf("%w: %w", closeProducerError{}, err)
		return err
	}

	log.Warning("producer rabbit is closed")
	return err
}

func (r *RabbitWorker) closeConn() error {
	var err error

	err = r.checkConn()
	if err != nil {
		err = fmt.Errorf("%w: %w", closeConnError{}, err)
		return err
	}

	err = r.streamEnv.Close()
	if err != nil {
		err = fmt.Errorf("%w: %w", closeConnError{}, err)
		return err
	}

	log.Warning("connect rabbit is closed")
	return err
}

func (r *RabbitWorker) Shutdown() {
	r.cancel()
	r.closeProducer()
	r.closeConn()
}

func (r *RabbitWorker) producerProcess(ctx context.Context) {
	var err error

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-r.incomingCh:
			message, ok := msg.(incomingMessage)
			if !ok {
				err = messageConversionError{"amqp.producerProcess"}
				log.Error(err)
				r.Shutdown()
				return
			}

			err = r.sendMessage(message.GetMessage())
			if err != nil {
				err = fmt.Errorf("%w: %w", messageSendError{}, err)
				message.Unsuccessfully()
				log.Error(err)
				r.errorHandling(err)
			} else {
				log.Info("message is delivered")
				message.Successfully()
			}

		}
	}
}

func (r *RabbitWorker) errorHandling(err error) {
	switch {
	case errors.Is(err, envStreamIError{}):
		r.createConn()
	case errors.Is(err, producerNotDefineError{}):
		r.createProducer()
	}
}
