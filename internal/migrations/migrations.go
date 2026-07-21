package migrations

import _ "embed"

//go:embed 001_initial.sql
var initialSQL string

func InitialSQL() string {
	return initialSQL
}
