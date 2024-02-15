package core

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func InitCore(limitBatch int) *Core {
	res := &Core{
		storageCh:  make(chan interface{}),
		limitBatch: limitBatch,
	}

	return res
}

func (c *Core) GetStorageCh() chan interface{} {
	return c.storageCh
}

type Core struct {
	storageCh  chan interface{}
	lastOffset int
	limitBatch int
	cancel     func()
}

func (c *Core) StartProcess() {
	ctx, cancel := context.WithCancel(context.Background())

	c.cancel = cancel

	go c.coreProcess(ctx)
}

func (c *Core) coreProcess(ctx context.Context) {
	offset, err := c.getLastOffset(10)

	if err != nil {
		return
	}

	c.lastOffset = offset

	for {
		err := c.workWithBatch(ctx)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (c *Core) workWithBatch(ctx context.Context) error {
	var offset int
	var err error

	defer func() {
		if offset > 0 {
			c.lastOffset = offset
		}
	}()

	// формирование запроса
	msg := transportMsg{
		typeMsg:   GetBatchMessages,
		param1:    c.lastOffset,
		param2:    c.limitBatch,
		reverseСh: make(chan interface{}),
	}

	// отправка сообщения в модуль storage
	c.storageCh <- msg

	for {
		select {
		case <-ctx.Done():
			// прекращение работы
			return err
		case answ, chOk := <-msg.reverseСh:
			answer, ok := answ.(AnswerStorage)
			if !ok {
				err = messageConversionError{"core.workWithBatch"}
				return err
			}

			err := answer.GetError()
			if err != nil {
				err = fmt.Errorf("%w: %w", storageError{}, err)
				return err
			}

			offset = answer.GetOffset()

			log.Infof("got offset: %d", offset)

			// прекращение работы, когда закрывается канал
			if !chOk {
				c.lastOffset = answer.GetOffset()
				return err
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func (c *Core) getLastOffset(timeOut int) (int, error) {
	var err error
	var offset int

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)

	msg := transportMsg{
		typeMsg:   GetLastOffset,
		reverseСh: make(chan interface{}),
	}

	c.storageCh <- msg

	select {
	case <-ctx.Done():
		// время ожидания вышло
		err = fmt.Errorf("%w: %w", timeOutError{}, noResponseStorageError{})
		log.Error(err)
		return offset, err
	case answ := <-msg.GetCh():
		answer, ok := answ.(AnswerStorage)
		if !ok {
			err = messageConversionError{"core.getLastOffset"}
			log.Error(err)
			return offset, err
		}

		err = answer.GetError()
		if err != nil {
			err = fmt.Errorf("%w: %w", storageError{}, err)
			log.Error(err)
			return offset, err
		}

		offset = answer.GetOffset()
		log.Info("offset received")
		return offset, err
	}
}

func (c *Core) Shutdown(err error) {
	c.cancel()
	log.Warning(err)
}

// метод для отправки сообщения
// func (c *Core) sendTransportMsg(msg transportMsg) error {
// 	c.storageCh <- msg

// }
