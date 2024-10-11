package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"redis-clone/redis_clone"
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
		go redis_clone.HandleRequest(conn)
		go redis_clone.Persist()
	}

}
