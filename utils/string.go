package utils

func Substr(str string, length int) string {
	var rslt []rune
	for idx, val := range str {
		if idx >= length {
			break
		}
		rslt = append(rslt, val)
	}
	return string(rslt)
}
