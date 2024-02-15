package psql

import "fmt"

type reconnectionAttemptsEndedError struct {
}

func (e reconnectionAttemptsEndedError) Error() string {
	return "the number of attempts to connect to PostgreSQL has ended"
}

type terminationPsqlError struct {
}

func (e terminationPsqlError) Error() string {
	return "psql stopped working for a reason ->"
}

type terminationPgConnError struct {
}

func (e terminationPgConnError) Error() string {
	return "pgConn stopped working for a reason ->"
}

type connIsNotReadyError struct {
}

func (e connIsNotReadyError) Error() string {
	return "connect PostgreSQL is not ready"
}

type messageConversionError struct {
	place string
}

func (e messageConversionError) Error() string {
	return fmt.Sprintf("message conversion error: %s", e.place)
}

type unknownTypeMessageError struct {
	message string
}

func (e unknownTypeMessageError) Error() string {
	return fmt.Sprintf("the type of message is unknown : %s", e.message)
}

type requestAttemptsEndedError struct {
	message string
}

func (e requestAttemptsEndedError) Error() string {
	return fmt.Sprintf("the number of request attempts has ended")
}
