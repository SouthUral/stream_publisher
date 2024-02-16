package core

import "fmt"

type messageConversionError struct {
	place string
}

func (e messageConversionError) Error() string {
	return fmt.Sprintf("message conversion error: %s", e.place)
}

type failedToSendMessageError struct {
}

func (e failedToSendMessageError) Error() string {
	return "failed to send message to stream"
}

type storageError struct {
}

func (e storageError) Error() string {
	return "error on the storage side"
}

type timeOutError struct {
}

func (e timeOutError) Error() string {
	return "time out error"
}

type noResponseStorageError struct {
}

func (e noResponseStorageError) Error() string {
	return "no response from storage"
}
