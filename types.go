package genroutine

import "genroutine/transaction"

// P - параметры, R - результат
type (
	// LoadDataList загрузить список по параметрам
	LoadDataList[P, R any] func(ts transaction.Session, param P) ([]R, error)
	// LoadData загрузить дату по параметрам
	LoadData[P, R any] func(ts transaction.Session, param P) (R, error)
	// ExecList выполнить список параметров
	ExecList[P any] func(ts transaction.Session, paramList []P) error
	// Exec выполнить параметры
	Exec[P any] func(ts transaction.Session, param P) error
)
