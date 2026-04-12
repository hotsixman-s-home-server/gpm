package util

func ArrayMap[T, U any](arr []T, callback func(value T, index int, array []T) U) []U {
	resultArr := make([]U, len(arr))
	for i, v := range arr {
		resultArr[i] = callback(v, i, arr)
	}
	return resultArr
}

func ArrayFilter[T any](arr []T, callback func(value T, index int, array []T) bool) []T {
	resultArr := make([]T, 0)
	for i, v := range arr {
		pass := callback(v, i, arr)
		if pass {
			resultArr = append(resultArr, v)
		}
	}
	return resultArr
}
