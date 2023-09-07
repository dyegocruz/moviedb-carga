package util

func ReverseArray(arrayElement []string) []string {
	arrayElementSize := len(arrayElement)
	revArr := make([]string, arrayElementSize)
	j := 0
	for i := arrayElementSize - 1; i >= 0; i-- {
		revArr[j] = arrayElement[i]
		j++
	}

	return revArr
}
