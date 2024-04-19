package redis_clone

import (
	"errors"
	"strconv"
	"sync"
)

var lock sync.Mutex
var data = map[string]string{}

func handleMutableOperation(op Operation) (string, error) {
	lock.Lock()
	defer lock.Unlock()
	switch op.Command {
	case SET:
		data[op.Key] = op.Value
		return "OK", nil
	case INCR:
		value, ok := data[op.Key]
		if !ok {
			data[op.Key] = "1"
			return "1", nil
		}
		i, err := strconv.Atoi(value)
		if err != nil {
			return "", errors.New("not integer")
		}
		i = i + 1
		value = strconv.Itoa(i)
		data[op.Key] = value
		return value, nil
	case DECR:
		value, ok := data[op.Key]
		if !ok {
			data[op.Key] = "-1"
			return "-1", nil
		}
		i, err := strconv.Atoi(value)
		if err != nil {
			return "", errors.New("not integer")
		}
		i = i - 1
		value = strconv.Itoa(i)
		data[op.Key] = value
		return value, nil
	default:
		return "", errors.New("invalid mutation command")
	}
}

func handleImmutableOperation(op Operation) (string, error) {
	switch op.Command {
	case GET:
		value, ok := data[op.Key]
		if !ok {
			return "(nil)", errors.New("key not found")
		}
		return value, nil
	case EXISTS:
		_, ok := data[op.Key]
		if !ok {
			return "0", nil
		} else {
			return "1", nil
		}
	default:
		return "PONG", nil // TODO
	}
}
