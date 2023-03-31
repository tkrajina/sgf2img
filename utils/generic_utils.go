package utils

func MinMax[T int | int32 | int64 | float32 | float64](t1, t2 T) (T, T) {
	if t1 < t2 {
		return t1, t2
	}
	return t2, t1
}
