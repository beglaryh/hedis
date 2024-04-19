package redis_clone

import "errors"

type TypeOfRequest int

const (
	SIMPLE_STRING = iota + 1
	STRING
	NUMBER
	ARRAY
)

func TypeOfRequestFrom(str string) (TypeOfRequest, error) {
	if str[0] == '*' {
		return ARRAY, nil
	}
	return TypeOfRequest(0), errors.New("unknown type of request")
}
