package util

import "strconv"

func ParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
