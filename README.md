# Специфичная либа для обхода некоторых ограничений оракловой либы go-oci8
## Пример
```
expectedData1 := []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5}

	// 5 раз конкурентно опросит бд и объеденит результаты в один слайс
	data1, err := genroutine.LoopReturnDataList(ctx, sm,
		func(ts transaction.Session, number int) (res []int, err error) {
			// подключаемся к бд, получаем результаты
			return []int{number, number}, nil
		}, []int{1, 2, 3, 4, 5})
	if err != nil {
		panic(err)
	}

	slices.Sort(data1)

	if !reflect.DeepEqual(expectedData1, data1) {
		panic("not equal")
	}

	// разделит paramList на равные части (согласно offset) - [1,2], [3,4], [5]
	// выполнит опрос бд 3 раза, в основном необходим для insert
	if err = genroutine.OffsetLoopReturnErr(ctx, sm, func(ts transaction.Session, paramList []int) error {
		return nil
	}, []int{1, 2, 3, 4, 5}, 2); err != nil {
		panic(err)
	}
```
