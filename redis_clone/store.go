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
		return handleIncrement(op.Key, "1")
	case INCRBY:
		return handleIncrement(op.Key, op.Value)
	case DECR:
		return handleIncrement(op.Key, "-1")
	case DECRBY:
		return handleDecrement(op.Key, op.Value)
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

func handleIncrement(key string, amount string) (string, error) {
	incrementAmount, err := strconv.Atoi(amount)
	if err != nil {
		return "", errors.New("did not provide a valid number")
	}
	value, ok := data[key]

	if !ok {
		data[key] = amount
		return amount, nil
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return "", errors.New("not integer")
	}
	i = i + incrementAmount
	value = strconv.Itoa(i)
	data[key] = value
	return value, nil
}

func handleDecrement(key, amount string) (string, error) {
	val, err := strconv.Atoi(amount)
	if err != nil {
		return "", errors.New("did not provide a valid number")
	}
	return handleIncrement(key, strconv.Itoa(-1*val))
}
