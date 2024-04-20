package redis_clone

import (
	"errors"
	"github.com/beglaryh/gocommon/collection"
	"strconv"
	"sync"
)

var lock sync.Mutex
var data = map[string]valueElement{}

func handleMutableOperation(op Operation) (string, error) {
	lock.Lock()
	defer lock.Unlock()
	switch op.Command {
	case SET:
		data[op.getKey()] = valueElement{v: op.getValue(), et: ESTRING}
		return "OK", nil
	case INCR:
		return handleIncrement(op.getKey(), "1")
	case INCRBY:
		return handleIncrement(op.getKey(), op.getValue())
	case DECR:
		return handleIncrement(op.getKey(), "-1")
	case DECRBY:
		return handleDecrement(op.getKey(), op.getValue())
	case RPUSH:
		return handlePush(op.getKey(), op.Values)
	default:
		return "", errors.New("invalid mutation command")
	}
}

func handleImmutableOperation(op Operation) (string, error) {
	switch op.Command {
	case GET:
		value, ok := data[op.getKey()]
		if !ok {
			return "(nil)", errors.New("key not found")
		} else if value.et != ESTRING {
			return "", errors.New("not returnable")
		}

		return value.v.(string), nil
	case EXISTS:
		_, ok := data[op.getKey()]
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
	if value.et != ESTRING {
		return "", errors.New("not integer")
	}

	if !ok {
		data[key] = valueElement{amount, ESTRING}
		return amount, nil
	}
	i, err := strconv.Atoi(value.v.(string))
	if err != nil {
		return "", errors.New("not integer")
	}
	i = i + incrementAmount
	v := strconv.Itoa(i)
	value.v = v
	data[key] = value
	return v, nil
}

func handleDecrement(key, amount string) (string, error) {
	val, err := strconv.Atoi(amount)
	if err != nil {
		return "", errors.New("did not provide a valid number")
	}
	return handleIncrement(key, strconv.Itoa(-1*val))
}

func handlePush(key string, values collection.List[string]) (string, error) {
	val, ok := data[key]
	if !ok {
		v := collection.NewLinkedList[string]()
		values.Stream().ForEach(func(e string) { _ = v.Add(e) })
		newVal := valueElement{v: v, et: ELIST}
		data[key] = newVal
	} else {
		list := val.v.(collection.LinkedList[string])
		values.Stream().ForEach(func(e string) { _ = list.Add(e) })
		newVal := valueElement{v: list, et: ELIST}
		data[key] = newVal
	}

	return strconv.Itoa(values.Size()), nil
}
