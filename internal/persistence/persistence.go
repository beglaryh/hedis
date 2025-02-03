package persistence

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/beglaryh/gocommon"
	"github.com/beglaryh/gocommon/collection/list/linkedlist"
	domain "github.com/beglaryh/hedis/internal/common"
	"github.com/beglaryh/hedis/internal/store"
)

var mutex sync.Mutex

func Persist(op domain.Operation) {
	mutex.Lock()
	defer mutex.Unlock()
	file, err := os.OpenFile("events.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Write to the file
	if _, err := file.WriteString(op.ToLog() + "*#\n"); err != nil {
		log.Fatal(err)
	}
}

func ReplayEvents() {
	file, err := os.Open("events.log")
	if err != nil {
		log.Println(err)
		return
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(splitter)

	// Loop over each line
	for scanner.Scan() {
		line := scanner.Text()
		op := toOperation(gocommon.String(line))
		store.HandleMutableOperation(op)
	}

	file.Close()
}

func toOperation(log gocommon.String) domain.Operation {
	split := log.Split("*#*")
	cmd, _ := strconv.Atoi(string(split[0]))
	keys := linkedlist.New[string]()
	values := linkedlist.New[string]()

	fmt.Println(log)
	fmt.Println(split)
	kvs := split[1].Split(",")
	vs := split[2].Split("*,*")

	for _, e := range kvs {
		keys.Add(string(e))
	}

	for _, e := range vs {
		values.Add(string(e))
	}

	return domain.Operation{
		Command: domain.Command(cmd),
		Keys:    &keys,
		Values:  &values,
	}
}

func splitter(data []byte, end bool) (int, []byte, error) {
	if end && len(data) == 0 {
		return 0, nil, nil
	}

	str := string(data)
	index := strings.Index(str, "*#\n")
	if index == -1 {
		return len(data), data, nil
	}
	token := str[0:index]
	return index + 3, []byte(token), nil
}
