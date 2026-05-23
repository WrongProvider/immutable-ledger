package main

import (
	"bufio"
	"fmt"
	"net"
)

const PORT = "8000" 
const PROTOCOL = "tcp"

func main () {
	ln, err := net.Listen(PROTOCOL, ":" + PORT)
	if err != nil {
		panic(err)
	}
	fmt.Println("Server running on :8000")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		go Handle(conn)
	}
}

func Handle(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected")
			return
		}
		fmt.Println("Received:", msg)
		conn.Write([]byte("Echo: " + msg))
	}
}

