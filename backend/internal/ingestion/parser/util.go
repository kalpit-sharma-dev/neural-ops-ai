package parser

import (
	"strconv"
)

func parseInt(value string) (int, error) {
	return strconv.Atoi(value)
}
