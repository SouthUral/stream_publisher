package psql

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// функция инициализирует подключение к PostgreSQL
func InitPsql(url string, timeWaitConn, timeWaitCheck, numAttempt, timeOut int) *psql {
	ctx, cancel := context.WithCancel(context.Background())

	conf := configPg{
		url:           url,
		timeWaitConn:  timeWaitConn,
		timeWaitCheck: timeWaitCheck,
		numAttempt:    numAttempt,
		timeOut:       timeOut,
	}

	pq := &psql{
		cancel:     cancel,
		conn:       initPgConn(conf),
		incomingCh: make(chan interface{}),
	}

	go pq.processReceivingEvents(ctx)

	return pq
}

type psql struct {
	cancel     func()
	conn       *pgConn
	mx         sync.RWMutex
	incomingCh chan interface{}
}

func (p *psql) GetCh() chan interface{} {
	return p.incomingCh
}

// процесс для получения события и ответа на них
func (p *psql) processReceivingEvents(ctx context.Context) {
	defer log.Warning("processReceivingEvents has finished executing")

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-p.incomingCh:
			message, ok := msg.(incomigMessage)
			if !ok {
				p.Shutdown(messageConversionError{"psql"})
				return
			}
			p.chooseAction(ctx, message)
		}
	}
}

// метод выбирает действия исходя из типа сообщения и запускает его
func (p *psql) chooseAction(ctx context.Context, message incomigMessage) {
	typeMsg := message.GetTypeMessage()
	switch typeMsg {
	case GetBatchMessages:
		go p.getBatchEvents(ctx, message)
	case GetLastOffset:
		go p.getLastOffset(ctx, message, 20, 2)
	case UpdateLastOffset:
		go p.updateLastOffset(ctx, message)
	case CleanLastOffset:
		go p.cleanLastOffset(ctx, message)
	default:
		log.Warning(unknownTypeMessageError{typeMsg})
	}
}

// процесс получения и отправки данных
func (p *psql) getBatchEvents(ctx context.Context, msg incomigMessage) {
	var p1, p2 int // временные параметры

	ch := msg.GetCh()

	p1 = msg.GetParamFirst()
	p2 = msg.GetParamSecond()

	rows, err := p.conn.RequestRowsEvents(ctx, gettingBundleOfEvents, p1, p2)

	if err != nil {
		log.Error(err)
		ch <- AnswerData{
			err: err,
		}
		return
	}

	numRows := len(rows)
	// отправка ответа
	for num, eventData := range rows {
		rowsLeft := numRows - num + 1
		ch <- AnswerData{
			rowsLeft: rowsLeft,
			data:     eventData,
		}
	}

	defer close(ch)
}

func (p *psql) getLastOffset(ctx context.Context, msg incomigMessage, numAttempt, timeWait int) {
	ch := msg.GetCh()

	offset, err := func() (int, error) {
		var offset int
		var err error
		for i := 0; i < numAttempt; i++ {
			offset, err = p.conn.RequestRowInt(ctx, getLastOffset)
			if err != nil {
				log.Warningf("retry the request %s", getLastOffset)
			} else {
				return offset, err
			}
			time.Sleep(time.Duration(timeWait) * time.Second)
		}
		err = fmt.Errorf("%w: %w", requestAttemptsEndedError{}, err)
		return offset, err
	}()

	if err != nil {
		ch <- AnswerData{
			err: err,
		}
		return
	}

	ch <- AnswerData{
		data: EventData{
			Offset: offset,
		},
	}

	defer close(ch)
}

func (p *psql) updateLastOffset(ctx context.Context, msg incomigMessage) {

	ch := msg.GetCh()

	p1 := msg.GetParamFirst()

	err := p.conn.UpdateOffset(ctx, updateOffset, p1)
	if err != nil {
		ch <- AnswerData{
			err: err,
		}
		return
	}

	ch <- AnswerData{}
	defer close(ch)
}

func (p *psql) cleanLastOffset(ctx context.Context, msg incomigMessage) {

	ch := msg.GetCh()

	err := p.conn.UpdateOffset(ctx, updateOffset, 0)
	if err != nil {
		ch <- AnswerData{
			err: err,
		}
		return
	}

	ch <- AnswerData{}
	defer close(ch)
}

// func (p *psql) makeRequest(someFunc func(context.Context, incomigMessage) error, numAttempt int) {
// 	err := f()

// }

// метод для прекращения работы psql
//   - err - ошибка, по причине которой была прекращена работа
func (p *psql) Shutdown(err error) {
	err = fmt.Errorf("%w: %w", terminationPsqlError{}, err)
	log.Warning(err)
	p.conn.Shutdown(err)

	p.cancel()
}
