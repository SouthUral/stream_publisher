package core

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
