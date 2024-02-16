package core

import "context"

type transportMsg struct {
	param1    int
	param2    int
	typeMsg   string
	reverseСh chan interface{}
}

func (t transportMsg) GetTypeMessage() string {
	return t.typeMsg

}

func (t transportMsg) GetParamFirst() int {
	return t.param1
}

func (t transportMsg) GetParamSecond() int {
	return t.param2
}

func (t transportMsg) GetCh() chan interface{} {
	return t.reverseСh
}

func initMessageToStream(message []byte) *messageToStream {
	ctxSuccessfull, cancelSuccessfull := context.WithCancel(context.Background())
	ctxUnSuccessfull, cancelUnSuccessfull := context.WithCancel(context.Background())

	res := &messageToStream{
		message:          message,
		ctxSuccessfully:  ctxSuccessfull,
		ctxUnSuccessfull: ctxUnSuccessfull,
		successfully:     cancelSuccessfull,
		unsuccessfully:   cancelUnSuccessfull,
	}

	return res
}

type messageToStream struct {
	message          []byte
	ctxSuccessfully  context.Context
	ctxUnSuccessfull context.Context
	successfully     func()
	unsuccessfully   func()
}

func (m messageToStream) GetMessage() []byte {
	return m.message
}

func (m messageToStream) Successfully() {
	m.successfully()
}

func (m messageToStream) Unsuccessfully() {
	m.unsuccessfully()
}

func (m messageToStream) GetSuccsessfull() <-chan struct{} {
	return m.ctxSuccessfully.Done()
}

func (m messageToStream) GetUnSuccsessfull() <-chan struct{} {
	return m.ctxUnSuccessfull.Done()
}
