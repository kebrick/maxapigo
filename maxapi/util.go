package maxapi

import "strconv"

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

func intToString(v int) string {
	return strconv.Itoa(v)
}

