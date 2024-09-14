package genroutine

// SessionManager обертка для коннекта к бд
// например: sqlx.DB
type SessionManager interface {
	CreateSession() Session
}

// Session обертка для транзакции к бд
// например: sqlx.Tx
type Session interface {
	Start() error
	Rollback() error
	Commit() error
}
