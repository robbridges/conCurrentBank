package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type TransactionType int

const (
	Deposit TransactionType = iota
	Withdrawal
)

type BankAccount struct {
	value               int
	mux                 sync.Mutex
	Posted              []Transaction
	Pending             []Transaction
	PendingTransactions chan Transaction
	PostedTransactions  chan Transaction
}

type Transaction struct {
	ID    uuid.UUID
	Value int
	Type  TransactionType
}

func main() {
	var wg sync.WaitGroup

	account := &BankAccount{
		value:               0,
		mux:                 sync.Mutex{},
		PendingTransactions: make(chan Transaction, 10), // Buffered channel to hold the transactions
		PostedTransactions:  make(chan Transaction, 10), // Buffered channel to hold the transactions
	}

	// Start a goroutine to process transactions
	wg.Add(1)
	go func() {
		defer wg.Done()
		account.processTransactions()
	}()

	// Start a goroutine to complete transactions
	wg.Add(1)
	go func() {
		defer wg.Done()
		account.completeTransaction()
	}()

	transactions := []struct {
		value           int
		transactionType TransactionType
	}{
		{100, Deposit},
		{50, Withdrawal},
		{20, Deposit},
	}

	// Add a deposit transaction
	for _, transaction := range transactions {
		wg.Add(1)
		go func(value int, ttype TransactionType) {
			defer wg.Done()
			account.addTransaction(value, ttype)
		}(transaction.value, transaction.transactionType)
	}

	// Wait for all goroutines to finish
	time.Sleep(2 * time.Second)
	close(account.PostedTransactions)
	close(account.PendingTransactions)
	wg.Wait()

	fmt.Println(account.getBalance())
	fmt.Println(account.Posted)

	fmt.Println("Press ENTER to exit...")
	fmt.Scanln()
}

func (b *BankAccount) deposit(amount int) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.value += amount
}

func (b *BankAccount) withdraw(amount int) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	tempAmount := b.value - amount
	if tempAmount < 0 {
		return errors.New("Withdrawal failed, insufficent funds")
	}
	b.value -= amount
	return nil
}

func (b *BankAccount) getBalance() int {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.value
}

func (b *BankAccount) addTransaction(value int, transactionType TransactionType) {
	b.mux.Lock()
	defer b.mux.Unlock()
	transaction := Transaction{ID: uuid.New(), Value: value, Type: transactionType}
	b.Pending = append(b.Pending, transaction) // Add transaction to Pending slice
	select {
	case b.PendingTransactions <- transaction:
		fmt.Printf("Added a %s transaction of value %d\n", transactionType, value)
	default:
		fmt.Println("Failed to add transaction: channel is full")
	}
}

func (b *BankAccount) processTransactions() {
	for {
		select {
		case transaction, ok := <-b.PendingTransactions:
			if ok {
				b.mux.Lock()
				b.Pending = removeTransaction(b.Pending, transaction.ID)
				switch transaction.Type {
				case Deposit:
					b.value += transaction.Value
					b.PostedTransactions <- transaction
					fmt.Printf("Processed a %s transaction of value %d\n", transaction.Type, transaction.Value)

				case Withdrawal:
					if b.value >= transaction.Value {
						b.value -= transaction.Value
						b.PostedTransactions <- transaction
						fmt.Printf("Processed a %s transaction of value %d\n", transaction.Type, transaction.Value)
					}
				}
				b.mux.Unlock()
			} else {
				fmt.Println("No transactions found to process")
				return
			}
		default:
			// No transaction is ready to be processed, so we can do other work here
		}
	}
}

func removeTransaction(transactions []Transaction, id uuid.UUID) []Transaction {
	for i, t := range transactions {
		if t.ID == id {
			return append(transactions[:i], transactions[i+1:]...)
		} else {
			fmt.Println("Could not find transaction to remove")
		}
	}
	return transactions
}

func (b *BankAccount) completeTransaction() {
	for transaction := range b.PostedTransactions {
		b.mux.Lock()
		b.Posted = append(b.Posted, transaction)
		fmt.Printf("Completed a %s transaction of value %d\n", transaction.Type, transaction.Value)
		b.mux.Unlock()
	}
}

func (t TransactionType) String() string {
	switch t {
	case Deposit:
		return "Deposit"
	case Withdrawal:
		return "Withdrawal"
	default:
		return "Unknown"
	}
}
