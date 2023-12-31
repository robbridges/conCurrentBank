package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
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
	transactionCount    int
}

type Transaction struct {
	ID    uuid.UUID
	Value int
	Type  TransactionType
}

func startBank(transactions []struct {
	value           int
	transactionType TransactionType
}) int {
	var wg sync.WaitGroup

	account := &BankAccount{
		value:               0,
		mux:                 sync.Mutex{},
		PendingTransactions: make(chan Transaction, 10), // Buffered channel to hold the transactions
		PostedTransactions:  make(chan Transaction, 10), // Buffered channel to hold the transactions
		transactionCount:    0,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := account.processTransactions()
		if err != nil {
			panic(err)
		}
		if account.transactionCount == len(transactions) {
			close(account.PostedTransactions)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		account.completeTransaction()
	}()

	// Start a single goroutine to add all transactions
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, transaction := range transactions {
			err := account.addTransaction(transaction.value, transaction.transactionType)
			if err != nil {
				continue
			}
		}
		close(account.PendingTransactions)
	}()
	wg.Wait()

	fmt.Println(account.getBalance())
	return account.getBalance()
}

func (b *BankAccount) deposit(amount int) {
	// no need for a mutex here, this is called within a function that is wrapped in a mutex
	b.value += amount
}

func (b *BankAccount) withdraw(amount int) error {
	// also no need for a mutex
	tempAmount := b.value - amount
	if tempAmount < 0 {
		return errors.New("withdrawal failed, insufficient funds")
	}
	b.value -= amount
	return nil
}

func (b *BankAccount) getBalance() int {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.value
}

func (b *BankAccount) addTransaction(value int, transactionType TransactionType) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	transaction := Transaction{ID: uuid.New(), Value: value, Type: transactionType}
	select {
	case b.PendingTransactions <- transaction:
		fmt.Printf("Added a %s transaction of value %d\n", transactionType, value)
		b.Pending = append(b.Pending, transaction)
	default:
		err := errors.New("unable to add transaction")
		return err
	}
	return nil
}

func (b *BankAccount) processTransactions() error {
	for {
		select {
		case transaction, ok := <-b.PendingTransactions:
			if ok {
				b.mux.Lock()
				b.Pending = removeTransaction(b.Pending, transaction.ID)
				switch transaction.Type {
				case Deposit:
					b.deposit(transaction.Value)
					b.PostedTransactions <- transaction
					fmt.Printf("Processed a %s transaction of value %d\n", transaction.Type, transaction.Value)

				case Withdrawal:
					err := b.withdraw(transaction.Value)
					if err != nil {
						inSufficientFundsErr := errors.New("insufficient funds to complete transaction")
						b.mux.Unlock()
						return inSufficientFundsErr
					}

					b.PostedTransactions <- transaction
					fmt.Printf("Processed a %s transaction of value %d\n", transaction.Type, transaction.Value)

				}
				b.transactionCount++
				b.mux.Unlock()
			} else {
				fmt.Println("No transactions found to process")
				return nil
			}
		default:
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
		b.Posted = append(b.Posted, transaction)
		fmt.Printf("Completed a %s transaction of value %d\n", transaction.Type, transaction.Value)
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
