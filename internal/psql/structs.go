package psql

import "sync"

// стурктура с потокобезопасным доступом к флагу
type connectionFlag struct {
	isConnect bool
	mx        sync.RWMutex
}

// функция инициализации connectionFlag
func initConnFlag() *connectionFlag {
	res := &connectionFlag{
		mx: sync.RWMutex{},
	}

	return res
}

func (c *connectionFlag) setIsConn(value bool) {
	c.mx.Lock()
	c.isConnect = value
	c.mx.Unlock()
}

func (c *connectionFlag) getIsConn() bool {
	c.mx.RLock()
	res := c.isConnect
	c.mx.RUnlock()
	return res
}

type configPg struct {
	url           string
	timeWaitConn  int // время ожидания при переподключении
	timeWaitCheck int // время ожидания между проверками коннекта
	numAttempt    int // количество попыток подключения (за один раз)
	timeOut       int // время ожидания коннекта
}
