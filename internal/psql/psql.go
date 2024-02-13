package psql

import (
	"context"
	"fmt"
	"sync"
	"time"

	pgx "github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

// функция инициализирует подключение к PostgreSQL
func initPsql(url string, timeWaitConn, timeWaitCheck, numAttempt, timeOut int) *psql {
	ctx, cancel := context.WithCancel(context.Background())

	conf := configPsql{
		timeWaitConn:  timeWaitConn,
		timeWaitCheck: timeWaitCheck,
		numAttempt:    numAttempt,
		timeOut:       timeOut,
	}

	pq := &psql{
		url:    url,
		cancel: cancel,
		isConn: initConnFlag(),
		cofig:  conf,
	}

	go pq.processConn(ctx)

	return pq
}

type psql struct {
	url        string
	cancel     func()
	conn       *pgx.Conn
	mx         sync.RWMutex
	isConn     *connectionFlag // флаг, показывающий есть ли коннект
	cofig      configPsql
	incomingCh chan interface{}
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
				p.Shutdown(messageConversionError{})
				return
			}
			p.chooseAction(message)
		}
	}
}

func (p *psql) chooseAction(message incomigMessage) {

}

// процесс контролирующий подключение к PostgreSQL
func (p *psql) processConn(ctx context.Context) {
	defer log.Warning("processConn has finished executing")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !p.isConn.getIsConn() {
				err := p.connAttempt(ctx, p.cofig.timeWaitConn, p.cofig.numAttempt, p.cofig.timeOut)
				if err != nil {
					p.Shutdown(err)
					return
				}

			}

			time.Sleep(time.Duration(p.cofig.timeWaitCheck) * time.Second)
		}
	}

}

// метод производит заданное количество попыток подключения, через заданный интервал:
//   - timeWait - время ожидания между попытками подключения (в секундах);
//   - numAttempt - количество попыток;
//   - timeOut - время ожидания коннекта (в секундах)
func (p *psql) connAttempt(ctx context.Context, timeWait, numAttempt, timeOut int) error {
	var err, errConn error
	for i := 0; i < numAttempt; i++ {
		select {
		case <-ctx.Done():
			return err
		default:
			p.checkConn()
			if !p.isConn.isConnect {
				errConn = p.connPg(timeOut)
				if errConn != nil {
					log.Error(errConn)
				}
			} else {
				return err
			}
			time.Sleep(time.Duration(timeWait) * time.Second)
		}
	}

	err = fmt.Errorf("%w: %w", reconnectionAttemptsEndedError{}, errConn)
	return err
}

// метод подключения к PostgreSQL
//   - timeOut - время ожидания коннекта (в секундах)
func (p *psql) connPg(timeOut int) error {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)

	p.conn, err = pgx.Connect(ctx, p.url)
	if err != nil {
		log.Error(err)
		p.isConn.setIsConn(false)
	} else {
		p.isConn.setIsConn(true)
	}

	return err
}

func (p *psql) checkConn() {
	if p.conn.IsClosed() {
		p.isConn.setIsConn(false)
	}
}

// метод для прекращения работы psql
//   - err - ошибка, по причине которой была прекращена работа
func (p *psql) Shutdown(err error) {
	err = fmt.Errorf("%w: %w", terminationPsqlError{}, err)
	log.Warning(err)

	p.cancel()
}
