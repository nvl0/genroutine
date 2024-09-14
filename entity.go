package genroutine

// option струтура для обработки результата или ошибки
type option[R any] struct {
	Result R
	Err    error
}

// Resolve метод, который вернет результат и ошибку для обработки
func (o option[R]) Resolve() (R, error) {
	return o.Result, o.Err
}

// newResult результат есть, ошибки нет
func newResult[R any](data R) option[R] {
	return option[R]{
		Result: data,
	}
}

// newError результата нет, ошибка есть
func newError[R any](err error) option[R] {
	return option[R]{
		Err: err,
	}
}
