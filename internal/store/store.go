package store

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/beglaryh/gocommon/collection/list"
	"github.com/beglaryh/gocommon/collection/list/linkedlist"
	"github.com/beglaryh/gocommon/stream"
	domain "github.com/beglaryh/hedis/internal/common"
)

var (
	lock sync.Mutex
	data = map[string]valueElement{}
)

func GetData() *map[string]valueElement {
	return &data
}

func HandleMutableOperation(op domain.Operation) (string, error) {
	lock.Lock()
	defer lock.Unlock()
	switch op.Command {
	case domain.APPEND:
		return handleAppend(op.GetKey(), op.GetValue())
	case domain.SET:
		data[op.GetKey()] = valueElement{v: op.GetValue(), et: ESTRING}
		return "OK", nil
	case domain.INCR:
		return handleIncrement(op.GetKey(), "1")
	case domain.INCRBY:
		return handleIncrement(op.GetKey(), op.GetValue())
	case domain.DECR:
		return handleIncrement(op.GetKey(), "-1")
	case domain.DECRBY:
		return handleDecrement(op.GetKey(), op.GetValue())
	case domain.RPUSH:
		return handlePush(op.GetKey(), op.Values)
	case domain.DEL:
		return handleDelete(op.Keys)
	default:
		return "", errors.New("invalid mutation command")
	}
}

func handleAppend(key, value string) (string, error) {
	v, exists := data[key]
	newValue := ""
	if exists {
		if v.et != ESTRING {
			return "", errors.New("invalid operation TODO")
		}
		existingValue := v.v.(string)
		newValue = existingValue + value
	} else {
		newValue = value
	}
	var v2 valueElement
	v2.v = newValue
	v2.et = ESTRING
	data[key] = v2
	return strconv.Itoa(len(newValue)), nil
}

func handleDelete(keys list.List[string]) (string, error) {
	deleteFunc := func(key string) int {
		_, exists := data[key]
		if !exists {
			return 0
		}
		delete(data, key)
		return 1
	}

	total := stream.Map[string, int](keys.Stream().Slice(), deleteFunc).
		Reduce(0, func(a, b int) int { return a + b })

	return strconv.Itoa(total), nil
}

func HandleImmutableOperation(op domain.Operation) (string, error) {
	switch op.Command {
	case domain.GET:
		value, ok := data[op.GetKey()]
		if !ok {
			return "(nil)", errors.New("key not found")
		} else if value.et != ESTRING {
			return "", errors.New("not returnable")
		}

		return value.v.(string), nil
	case domain.EXISTS:
		_, ok := data[op.GetKey()]
		if !ok {
			return "0", nil
		} else {
			return "1", nil
		}
	case domain.LRANGE:
		s1, _ := op.Values.Get(0)
		s2, _ := op.Values.Get(1)
		start, _ := strconv.Atoi(s1)
		end, _ := strconv.Atoi(s2)
		return handleLRange(op.GetKey(), start, end)
	case domain.ECHO:
		return op.GetKey(), nil
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

func handlePush(key string, values list.List[string]) (string, error) {
	val, ok := data[key]
	if !ok {
		v := linkedlist.New[string]()
		values.Stream().ForEach(func(e string) { _ = v.Add(e) })
		newVal := valueElement{v: v, et: ELIST}
		data[key] = newVal
	} else {
		list := val.v.(linkedlist.LinkedList[string])
		values.Stream().ForEach(func(e string) { _ = list.Add(e) })
		newVal := valueElement{v: list, et: ELIST}
		data[key] = newVal
	}

	return strconv.Itoa(values.Size()), nil
}

// TODO for now handling basic range
func handleLRange(key string, start int, end int) (string, error) {
	val, ok := data[key]
	if !ok {
		return "", errors.New("TODO ERROR")
	}
	if val.et != ELIST {
		return "", errors.New("TODO ERROR")
	}
	ll := val.v.(linkedlist.LinkedList[string])
	arr := ll.ToArray()
	fmt.Println(end)

	if end > -1 && end <= len(arr) {
		end += 1
	} else if end < 0 {
		end = len(arr) + end + 1
		if end > len(arr) {
			end = end % len(arr)
		}
	}
	sub := arr[start:end]

	length := len(sub)
	format := "*%d\r\n%s"

	arrayStr := ""
	for _, e := range sub {
		s := fmt.Sprintf("$%d\r\n%s\r\n", len(e), e)
		arrayStr += s
	}
	response := fmt.Sprintf(format, length, arrayStr)
	return response, nil
}
