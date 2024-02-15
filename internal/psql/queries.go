package psql

const (
	gettingBundleOfEvents = `
		select
			id_offset,
			message
		from test_system.messages
		where id_offset > $1
		order by id_offset
		limit $2`
	getLastOffset = "SELECT test_system.get_last_offset()"
	updateOffset  = "CALL test_system.update_last_offset($1)"
)

// "SELECT test_system.get_messages($1, $2)"
