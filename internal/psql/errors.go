package psql

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

type messageConversionError struct {
}

func (e messageConversionError) Error() string {
	return "message conversion error"
}
