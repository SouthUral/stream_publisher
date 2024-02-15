package psql

import (
	"context"
	"fmt"
	"sync"
	"time"

	pgx "github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

type pgConn struct {
	url    string
	mx     sync.RWMutex
	conn   *pgx.Conn
	isConn *connectionFlag
	config configPg
	cancel func()
}

func initPgConn(config configPg) *pgConn {
	ctx, cancel := context.WithCancel(context.Background())

	res := &pgConn{
		url:    config.url,
		mx:     sync.RWMutex{},
		isConn: initConnFlag(),
		config: config,
		cancel: cancel,
	}

	go res.processConn(ctx)

	return res
}

// процесс контролирующий подключение к PostgreSQL
func (p *pgConn) processConn(ctx context.Context) {
	defer log.Warning("processConn has finished executing")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !p.isConn.getIsConn() {
				err := p.connAttempt(ctx, p.config.timeWaitConn, p.config.numAttempt, p.config.timeOut)
				if err != nil {
					p.Shutdown(err)
					return
				}

			}

			time.Sleep(time.Duration(p.config.timeWaitCheck) * time.Second)
		}
	}
}

// метод производит заданное количество попыток подключения, через заданный интервал:
//   - timeWait - время ожидания между попытками подключения (в секундах);
//   - numAttempt - количество попыток;
//   - timeOut - время ожидания коннекта (в секундах)
func (p *pgConn) connAttempt(ctx context.Context, timeWait, numAttempt, timeOut int) error {
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
func (p *pgConn) connPg(timeOut int) error {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)

	p.mx.Lock()
	p.conn, err = pgx.Connect(ctx, p.url)
	p.mx.Unlock()

	if err != nil {
		log.Error(err)
		p.isConn.setIsConn(false)
	} else {
		p.isConn.setIsConn(true)
		log.Info("Коннект состоялся")
	}

	return err
}

func (p *pgConn) checkConn() {
	var isClosed bool

	p.mx.Lock()
	if p.conn != nil {
		isClosed = p.conn.IsClosed()
	}
	defer p.mx.Unlock()

	if isClosed {
		p.isConn.setIsConn(false)
	}
}

func (p *pgConn) closeConn() {
	p.checkConn()
	if !p.isConn.getIsConn() {
		return
	}

	timeOut := time.Duration(5)
	ctx, _ := context.WithTimeout(context.Background(), timeOut*time.Second)

	p.mx.Lock()
	err := p.conn.Close(ctx)
	p.mx.Unlock()

	if err != nil {
		log.Error(err)
	}
}

// - err - ошибка, по причине которой была прекращена работа
func (p *pgConn) Shutdown(err error) {
	err = fmt.Errorf("%w: %w", terminationPgConnError{}, err)
	log.Warning(err)
	p.closeConn()

	p.cancel()
}

// метод для забора данных из БД
func (p *pgConn) RequestRowsEvents(ctx context.Context, query string, p1, p2 int) ([]EventData, error) {
	var err error
	var data []EventData

	if !p.isConn.getIsConn() {
		err = connIsNotReadyError{}
		log.Error(err)
		return data, err
	}

	log.Infof("pgconn.RequestRowsEvents p1: %d, p2: %d", p1, p2)
	p.mx.Lock()
	rows, err := p.conn.Query(ctx, query, p1, p2)

	if err != nil {
		log.Error(err)
		return data, err
	}
	defer p.mx.Unlock()
	defer rows.Close()

	data, err = pgx.CollectRows[EventData](rows, pgx.RowToStructByName[EventData])
	return data, err
}

// метод для забора одинарных int данных по запросу без параметров
func (p *pgConn) RequestRowInt(ctx context.Context, query string) (int, error) {
	var res int
	var err error

	if !p.isConn.getIsConn() {
		err = connIsNotReadyError{}
		log.Error(err)
		return res, err
	}

	p.mx.Lock()
	err = p.conn.QueryRow(ctx, query).Scan(&res)
	if err != nil {
		log.Error(err)
		return res, err
	}
	defer p.mx.Unlock()

	return res, err
}

func (p *pgConn) UpdateOffset(ctx context.Context, query string, offset int) error {
	var err error

	if !p.isConn.getIsConn() {
		err = connIsNotReadyError{}
		log.Error(err)
		return err
	}

	p.mx.Lock()
	_, err = p.conn.Exec(ctx, query, offset)
	if err != nil {
		log.Error(err)
	}
	defer p.mx.Unlock()

	return err
}
