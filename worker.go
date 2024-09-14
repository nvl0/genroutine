package genroutine

import (
	"fmt"
	"sync"
)

// returnDataWithRollback возвращает результат и выполняет роллбэк.
// Типы: P - параметры, R - результат.
func returnDataWithRollback[P, R any](sm SessionManager, f LoadData[P, R],
	param P, ch chan<- option[R], wg *sync.WaitGroup) {

	defer wg.Done()

	var (
		res R
		err error
	)

	ts := sm.CreateSession()
	if err = ts.Start(); err != nil {
		ch <- newError[R](fmt.Errorf("не удалось открыть транзакцию, ошибка %v", err))
		return
	}
	defer ts.Rollback()

	if res, err = f(ts, param); err != nil {
		ch <- newError[R](err)
		return
	}

	if err = ts.Rollback(); err != nil {
		ch <- newError[R](fmt.Errorf("не удалось закрыть транзакцию, ошибка %v", err))
		return
	}

	ch <- newResult(res)
}

// returnErrWithCommit возвращает ошибку и выполняет коммит.
// Тип P - параметры.
func returnErrWithCommit[P any](sm SessionManager, f Exec[P],
	param P, ch chan<- error, wg *sync.WaitGroup) {

	defer wg.Done()

	var (
		err error
	)

	ts := sm.CreateSession()
	if err = ts.Start(); err != nil {
		ch <- fmt.Errorf("не удалось открыть транзакцию, ошибка: %v", err)
		return
	}
	defer ts.Rollback()

	if err = f(ts, param); err != nil {
		ch <- err
		return
	}

	if err = ts.Commit(); err != nil {
		ch <- fmt.Errorf("не удалось закрыть транзакцию, ошибка: %v", err)
		return
	}

	ch <- nil
}
