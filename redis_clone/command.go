package redis_clone

import (
	"errors"
	"strings"

	"github.com/beglaryh/gocommon/collection/list"
	"github.com/beglaryh/gocommon/collection/list/linkedlist"
)

type Command int

const (
	APPEND Command = iota + 1
	ECHO
	COMMAND
	PING
	GET
	SET
	DEL
	INCR
	INCRBY
	DECR
	DECRBY
	RPUSH
	LPUSH
	LRANGE
	EXISTS
	CONFIG
)

func CommandFrom(str string) (Command, error) {
	str = strings.ToUpper(str)
	switch str {
	case "APPEND":
		return APPEND, nil
	case "SET":
		return SET, nil
	case "INCR":
		return INCR, nil
	case "INCRBY":
		return INCRBY, nil
	case "DECR":
		return DECR, nil
	case "DEL":
		return DEL, nil
	case "DECRBY":
		return DECRBY, nil
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
	case "RPUSH":
		return RPUSH, nil
	case "LPUSH":
		return LPUSH, nil
	case "LRANGE":
		return LRANGE, nil
	default:
		return Command(0), errors.New("unable to parse Command")
	}
}

func (comm Command) isMutation() bool {
	switch comm {
	case APPEND, SET, INCR, INCRBY, DECR, DECRBY, RPUSH, LPUSH, DEL:
		return true
	default:
		return false
	}
}

func (comm Command) hasIntegerResponse() bool {
	switch comm {
	case APPEND, INCR, INCRBY, DECR, DECRBY, EXISTS, RPUSH, LPUSH, DEL:
		return true
	default:
		return false
	}
}

func (comm Command) passThroughResponse() bool {
	return comm == LRANGE
}

func (comm Command) hasMultipleKeys() bool {
	switch comm {
	case DEL:
		return true
	default:
		return false
	}
}

func (comm Command) hasMultipleValues() bool {
	switch comm {
	case RPUSH, LPUSH, LRANGE:
		return true
	default:
		return false
	}
}

func (comm Command) hasValue() bool {
	switch comm {
	case APPEND, SET, INCRBY, DECRBY, RPUSH, LPUSH, CONFIG, LRANGE:
		return true
	default:
		return false
	}
}

type Operation struct {
	Command Command
	Keys    list.List[string]
	Values  list.List[string]
}

func newOperation() Operation {
	keysLL := linkedlist.New[string]()
	valuesLL := linkedlist.New[string]()
	keys := (list.List[string])(&keysLL)
	values := (list.List[string])(&valuesLL)
	return Operation{
		Command: Command(0),
		Keys:    keys,
		Values:  values,
	}
}

func (op Operation) getKey() string {
	key, _ := op.Keys.Get(0)
	return key
}

func (op Operation) getValue() string {
	value, _ := op.Values.Get(0)
	return value
}
