package psql

type EventData struct {
	Offset  int    `db:"id_offset"`
	Message []byte `db:"message"`
}

type AnswerData struct {
	rowsLeft int
	data     EventData
	err      error
}

func (a AnswerData) GetRowsLeft() int {
	return a.rowsLeft
}

func (a AnswerData) GetError() error {
	return a.err
}

func (a AnswerData) GetOffset() int {
	return a.data.Offset
}

func (a AnswerData) GetMessage() []byte {
	return a.data.Message
}
