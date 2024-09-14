package genroutine

import (
	"context"
	"errors"
	"genroutine/transaction"
	"math"
	"sync"
)

// LoopReturnDataList проходит цикл, возвращает результат, выполняет роллбэк.
// Подходит для множественного получения списка результатов. Типы: P - параметры, R - результат.
func LoopReturnDataList[P, R any](ctx context.Context, sm transaction.SessionManager, f LoadDataList[P, R],
	paramList []P) (res []R, err error) {
	res = make([]R, 0)

	var (
		iter = len(paramList)
		ch   = make(chan option[[]R], iter)
		wg   = sync.WaitGroup{}
	)
	wg.Add(iter)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, param := range paramList {
		// param = p, вторичный нейминг p нужен для избежания внутренней коллизии
		go returnDataWithRollback(sm, func(ts transaction.Session, p P) ([]R, error) {
			return f(ts, p)
		}, param, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

Loop:
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case opt, open := <-ch:
			if !open {
				break Loop
			}

			var r []R

			if r, err = opt.Resolve(); err != nil {
				return
			}

			res = append(res, r...)
		}
	}
	return
}

// OffsetLoopReturnErr проходит по циклу со сдвигом, возвращает ошибку, выполняет коммит.
// Подходит для множественной записи списков. Тип P - параметры.
func OffsetLoopReturnErr[P any](ctx context.Context, sm transaction.SessionManager, f ExecList[P],
	paramList []P, offset int) (err error) {
	if offset <= 0 {
		return errors.New("offset должен быть положительным")
	}

	var (
		iter = int(math.Ceil(float64(len(paramList)) / float64(offset)))
		ch   = make(chan error, iter)
		wg   = sync.WaitGroup{}
	)
	wg.Add(iter)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		start, end = 0, 0
		total      = len(paramList)
	)

	for start = 0; start < total; start += offset {
		end = start + offset
		if end > total {
			end = total
		}

		offsetParamList := paramList[start:end]

		// offsetParamList = opl, вторичный нейминг opl нужен для избежания внутренней коллизии
		go returnErrWithCommit(sm, func(ts transaction.Session, opl []P) error {
			return f(ts, opl)
		}, offsetParamList, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var open bool
Loop:
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case err, open = <-ch:
			if !open {
				break Loop
			}

			if err != nil {
				return
			}
		}
	}
	return
}
