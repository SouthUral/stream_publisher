package core

type AnswerStorage interface {
	GetRowsLeft() int
	GetError() error
	GetOffset() int
	GetMessage() []byte
}
