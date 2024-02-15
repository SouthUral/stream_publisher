package psql

type incomigMessage interface {
	GetTypeMessage() string
	GetParamFirst() int
	GetParamSecond() int
	GetCh() chan interface{}
}
