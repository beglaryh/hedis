package redis_clone

import (
	"errors"
	"strings"
)

type Command int

const (
	ECHO Command = iota + 1
	COMMAND
	PING
	GET
	SET
	INCR
	DECR
	EXISTS
	CONFIG
)

func CommandFrom(str string) (Command, error) {
	str = strings.ToUpper(str)
	switch str {
	case "SET":
		return SET, nil
	case "INCR":
		return INCR, nil
	case "DECR":
		return DECR, nil
	case "COMMAND":
		return COMMAND, nil
	case "GET":
		return GET, nil
	case "PING":
		return PING, nil
	case "ECHO":
		return ECHO, nil
	case "EXISTS":
		return EXISTS, nil
	case "CONFIG":
		return CONFIG, nil
	default:
		return Command(0), errors.New("unable to parse Command")
	}
}

func (comm Command) isMutation() bool {
	switch comm {
	case SET, INCR, DECR:
		return true
	default:
		return false
	}
}

func (comm Command) hasIntegerResponse() bool {
	switch comm {
	case INCR, DECR, EXISTS:
		return true
	default:
		return false
	}
}

type Operation struct {
	Command Command
	Key     string
	Value   string
}
