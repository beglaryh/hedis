package redis_clone

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	conn *net.Conn
}

const (
	ErrorFmt      = "-%s\r\n"
	IntegerFmt    = ":%s\r\n"
	SimpleFmt     = "+%s\r\n"
	BulkStringFmt = "$%d\r\n%s\r\n"
)

type requestState struct {
	start         bool
	totalElements int
	totalKeys     int
	totalValues   int
	elementCount  int
	op            Operation
}

func newRequestState() requestState {
	return requestState{
		start:     true,
		totalKeys: 1,
		op:        newOperation(),
	}
}

func (state *requestState) reset() {
	state.start = true
	state.totalKeys = 1
	state.totalElements = 0
	state.elementCount = 0
	state.op.Keys.Clear()
	state.op.Values.Clear()
	state.op.Command = Command(0)
}

func (state *requestState) requiresKey() bool {
	return !state.start && state.totalKeys != state.op.Keys.Size()
}

func (state *requestState) requiresValues() bool {
	return !state.start && state.totalValues != state.op.Values.Size()
}

func (state *requestState) addKey(key string) {
	_ = state.op.Keys.Add(key)
}

func (state *requestState) addValue(val string) {
	_ = state.op.Values.Add(val)
}

func (state *requestState) incrementCount() {
	state.elementCount += 1
}

func (state *requestState) readyToBeHandled() bool {
	return !state.start && state.totalElements == (state.op.Keys.Size()+state.op.Values.Size()+1)
}

func HandleRequest(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Split(splitter)
	state := newRequestState()
	var response string
	for scanner.Scan() {
		part := scanner.Text()
		if state.start {
			state.start = false
			if part[0] != '*' {
				response = fmt.Sprintf(ErrorFmt, "bad request")
				conn.Write([]byte(response))
				state.reset()
			} else {
				total, _ := strconv.Atoi(part[1:])
				state.totalElements = total
			}
			continue
		} else if part[0] == '$' {
			continue // TODO
		} else if state.op.Command == Command(0) {
			comm, err := CommandFrom(part)
			if err != nil {
				message := fmt.Sprintf("Err unknown command '%s'", part)
				response = fmt.Sprintf(ErrorFmt, message)
				_, err := conn.Write([]byte(response))
				if err != nil {
					return
				}
				state.reset()
			} else {
				state.op.Command = comm
				state.incrementCount()
				if comm.hasValue() {
					if comm.hasMultipleValues() {
						state.totalValues = state.totalElements - 2
					} else {
						state.totalValues = 1
					}
				} else if comm.hasMultipleKeys() {
					state.totalKeys = state.totalElements - 1
				}
			}
		} else if state.requiresKey() {
			state.addKey(part)
		} else if state.requiresValues() {
			state.addValue(part)
		}

		if state.readyToBeHandled() {

			if state.op.Command.isMutation() {
				resp, err := handleMutableOperation(state.op)
				if err != nil {
					response = fmt.Sprintf(ErrorFmt, err.Error())
				} else if state.op.Command.hasIntegerResponse() {
					response = fmt.Sprintf(IntegerFmt, resp)
				} else {
					response = fmt.Sprintf(BulkStringFmt, len(resp), resp)
				}
			} else {
				resp, err := handleImmutableOperation(state.op)
				if err != nil {
					if resp != "" {
						response = fmt.Sprintf(SimpleFmt, resp)
					} else {
						response = fmt.Sprintf(ErrorFmt, err.Error())
					}
				} else if state.op.Command.hasIntegerResponse() {
					response = fmt.Sprintf(IntegerFmt, resp)
				} else if state.op.Command.passThroughResponse() {
					response = resp
				} else {
					response = fmt.Sprintf(BulkStringFmt, len(resp), resp)
				}
			}
			state.reset()
			_, err := conn.Write([]byte(response))
			response = ""
			if err != nil {
				return
			}
		}
	}
}

func splitter(data []byte, end bool) (int, []byte, error) {
	str := string(data)
	if end {
		return len(data), data, nil
	}

	index := strings.Index(str, "\r\n")
	if index == -1 {
		return len(data), data, nil
	}
	token := str[0:index]
	return index + 2, []byte(token), nil
}
