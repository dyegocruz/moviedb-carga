package util

import "strconv"

func ArrayContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func StringToInt(val string) int {
	ret, err := strconv.Atoi(val)
	if err != nil {
		panic("Não foi possível converter a string para int")
	}
	return ret
}
