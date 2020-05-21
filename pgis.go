package pgis

// this is here so we can pass both sql.Row and sql.Rows to the
// PgisQueryRowToPgisRowFunc below (20170824/thisisaaronland)

type PgisResultSet interface {
	Scan(dest ...interface{}) error
}

type PgisResult interface {
	Row() interface{}
}

// these are badly named...

type PgisQueryRowFunc func(PgisResultSet) (PgisResult, error)

type PgisQueryResultFunc func(PgisResult, chan bool, chan error) error
