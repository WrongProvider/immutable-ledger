package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	_shutdownPeriod      = 15 * time.Second
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
	port                 = "8000"
	protocol             = "tcp"
	ip                   = "127.0.0.1"
)

type Transaction struct {
	TransactionId     int    `json:"transaction_id"`
	Value             int    `json:"value"`
	Origim            string `json:"origin"`
	Destination       string `json:"destination"`
	BankAccountNumber string `json:"bank_account_number"`
}

type Balance struct {
	BankAccountNumber int
	NumDeposits       int
	NumWithdraws      int
	TotalBalance      int
}

var isShuttingDown atomic.Bool

func main() {

	// setup signal context
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// endpoints
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if isShuttingDown.Load() {
			http.Error(w, "Shutting Down", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "OK")

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			fmt.Fprint(w, "Welcome to your ledger!")
		case <-r.Context().Done():
			http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		}
	})

	// Ensure in-flight requests aren't cancelled immediately on SIGTERM
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: ip + ":" + port,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	go func() {
		slog.Info("Server starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	json_msg := Transaction{1, 1000, "lucas", "matheus", "123-bank-id"}
	fmt.Println(GenerateHashFromTransaction(json_msg))

	// Wait for signal
	<-rootCtx.Done()
	stop()
	isShuttingDown.Store(true)
	log.Println("Received shutdown signal, shutting down.")

	// Give time for readiness check to propagate
	time.Sleep(_readinessDrainDelay)
	log.Println("Readiness check propagated, now waiting for ongoing requests to finish.")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()
	err := server.Shutdown(shutdownCtx)
	stopOngoingGracefully()
	if err != nil {
		log.Println("Failed to wait for ongoing requests to finish, waiting for forced cancellation.")
		time.Sleep(_shutdownHardPeriod)
	}

	log.Println("Server shut down gracefully.")

}

func Handle(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var t Transaction

	if err := decoder.Decode(&t); err != nil {
		slog.Error("fail to decode transaction", "error", err)
		return
	}

	if t.Value <= 0 {
		conn.Write([]byte("ERRO: valor invalido\n"))
		return
	}

	hash, _ := GenerateHashFromTransaction(t)
	slog.Info("transaction processed", "hash", hash)

	conn.Write([]byte("OK " + hash + "\n"))

}

func GenerateHashFromTransaction(t Transaction) (string, error) {
	serialized_transaction, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("Can't serialize transaction to hash %s: %w", t.TransactionId, err)
	}

	hash_transaction := sha256.Sum256(serialized_transaction)
	encoded_hash := hex.EncodeToString(hash_transaction[:])

	return encoded_hash, nil
}
