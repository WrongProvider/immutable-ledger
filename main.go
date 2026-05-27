package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
)

const PORT = "8000" 
const PROTOCOL = "tcp"

type Transaction struct {
	Value int
	Origim string
	Destination string
	BankAccountNumber string
}

type Balance struct {
	BankAccountNumber int
	NumDeposits int
	NumWithdraws int
	TotalBalance int

}
func main () {
	ln, err := net.Listen(PROTOCOL, ":" + PORT)
	if err != nil {
		panic(err)
	}
	fmt.Println("Server running on :8000")
	
	json_msg := Transaction{1000, "lucas", "matheus", "123-bank-id"}
	fmt.Println("without")
	fmt.Println(json_msg)
	fmt.Println(GenerateHashFromTransaction(json_msg))

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

func GenerateHashFromTransaction(t Transaction) string {
	serialized_transaction, err := json.Marshal(t)
	if err != nil { 
		fmt.Println("Can't serialize", serialized_transaction)
	}

	hash_transaction := sha256.Sum256(serialized_transaction)
	encoded_hash := hex.EncodeToString(hash_transaction)

	return string(encoded_hash)
}
	

