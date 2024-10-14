package main

import (
	"fmt"
	hedis "hedis/pkg"
	"log"
	"net"
	"os"
)

func main() {
	f, _ := os.ReadFile("splash.txt")
	fmt.Println(string(f))

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", "127.0.0.1", "6379"))
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go hedis.HandleRequest(conn)
		go hedis.Persist()
	}

}
