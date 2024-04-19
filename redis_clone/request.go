package redis_clone

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
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
	elementCount  int
}

func newRequestState() requestState {
	return requestState{start: true}
}

func (state *requestState) reset() {
	state.start = true
	state.totalElements = 0
	state.elementCount = 0
}

func (state *requestState) incrementCount() {
	state.elementCount += 1
}

func (state *requestState) readyToBeHandled() bool {
	return !state.start && state.totalElements == state.elementCount
}

func HandleRequest(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	state := newRequestState()
	var response string
	var operation Operation
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

		} else if part[0] == '$' {
			continue // TODO
		} else if state.elementCount == 0 {
			comm, err := CommandFrom(part)
			if err != nil {
				response = fmt.Sprintf("Err unknown command '%s'", part)
				_, err := conn.Write([]byte(response))
				if err != nil {
					return
				}
				state.reset()
			} else {
				operation.Command = comm
				state.incrementCount()
			}
		} else if state.elementCount == 1 {
			state.incrementCount()
			operation.Key = part
		} else if state.elementCount == 2 {
			state.incrementCount()
			operation.Value = part
		}

		if state.readyToBeHandled() {
			state.reset()
			if operation.Command.isMutation() {
				resp, err := handleMutableOperation(operation)
				if err != nil {
					response = fmt.Sprintf(ErrorFmt, err.Error())
				} else if operation.Command.hasIntegerResponse() {
					response = fmt.Sprintf(IntegerFmt, resp)
				} else {
					response = fmt.Sprintf(BulkStringFmt, len(resp), resp)
				}
			} else {
				resp, err := handleImmutableOperation(operation)
				if err != nil {
					if resp != "" {
						response = fmt.Sprintf(SimpleFmt, resp)
					} else {
						response = fmt.Sprintf(ErrorFmt, err.Error())
					}
				} else if operation.Command.hasIntegerResponse() {
					response = fmt.Sprintf(IntegerFmt, resp)
				} else {
					response = fmt.Sprintf(BulkStringFmt, len(resp), resp)
				}
			}
			_, err := conn.Write([]byte(response))
			if err != nil {
				return
			}
		}
	}
}
