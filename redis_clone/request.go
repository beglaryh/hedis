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

func HandleRequest(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	var response string
	var totalElements int
	var elementCount int
	var operation Operation
	var start = true
	for scanner.Scan() {
		part := scanner.Text()
		if start {
			if part[0] != '*' {
				response = fmt.Sprintf(ErrorFmt, "bad request")
				conn.Write([]byte(response))
			} else {
				total, _ := strconv.Atoi(part[1:])
				totalElements = total
			}
			start = false
		} else if part[0] == '$' {
			continue // TODO
		} else if elementCount == 0 {
			comm, err := CommandFrom(part)
			if err != nil {
				response = fmt.Sprintf("Err unknown command '%s'", part)
				conn.Write([]byte(response))
			}
			operation.Command = comm
			elementCount += 1
		} else if elementCount == 1 {
			elementCount += 1
			operation.Key = part
		} else if elementCount == 2 {
			elementCount += 1
			operation.Value = part
		}

		if elementCount == totalElements {
			start = true
			elementCount = 0
			totalElements = 0
			if operation.Command.isMutation() {
				resp, err := handleMutableOperation(operation)
				if err != nil {
					response = fmt.Sprintf(ErrorFmt, err.Error())
					conn.Write([]byte(response))
				} else if operation.Command.hasIntegerResponse() {
					response = fmt.Sprintf(IntegerFmt, resp)
					conn.Write([]byte(response))
				} else {
					response = fmt.Sprintf(BulkStringFmt, len(resp), resp)
					conn.Write([]byte(response))
				}
			} else {
				resp, err := handleImmutableOperation(operation)
				if err != nil {
					if resp != "" {
						response = fmt.Sprintf(SimpleFmt, resp)
					} else {
						response = fmt.Sprintf(ErrorFmt, err.Error())
					}
					conn.Write([]byte(response))
				} else if operation.Command.hasIntegerResponse() {
					response = fmt.Sprintf(IntegerFmt, resp)
					conn.Write([]byte(response))
				} else {
					response = fmt.Sprintf(BulkStringFmt, len(resp), resp)
					conn.Write([]byte(response))
				}
			}
		}
	}
}
