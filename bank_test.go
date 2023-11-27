package main

import (
	"sync"
	"testing"
)

func TestDeposit(t *testing.T) {
	account := &BankAccount{
		value:               0,
		mux:                 sync.Mutex{},
		PendingTransactions: make(chan Transaction, 10), // Buffered channel to hold the transactions
		PostedTransactions:  make(chan Transaction, 10), // Buffered channel to hold the transactions
		transactionCount:    0,
	}

	account.deposit(30)
	want := 30
	got := account.getBalance()
	if got != want {
		t.Errorf("Expected account to have a value of %d, got %d", want, got)
	}
}
func TestWithdraw(t *testing.T) {
	account := &BankAccount{
		value:               50,
		mux:                 sync.Mutex{},
		PendingTransactions: make(chan Transaction, 10), // Buffered channel to hold the transactions
		PostedTransactions:  make(chan Transaction, 10), // Buffered channel to hold the transactions
		transactionCount:    0,
	}
	account.withdraw(35)
	want := 15
	got := account.getBalance()
	if got != want {
		t.Errorf("Expected balance to be %d, got %d", want, got)
	}
}

func TestGetBalance(t *testing.T) {
	account := &BankAccount{
		value:               72,
		mux:                 sync.Mutex{},
		PendingTransactions: make(chan Transaction, 10), // Buffered channel to hold the transactions
		PostedTransactions:  make(chan Transaction, 10), // Buffered channel to hold the transactions
		transactionCount:    0,
	}
	want := 72
	got := account.getBalance()
	if got != want {
		t.Errorf("Unexpected balance returned, wanted %d, got %d", want, got)
	}

}

func TestAddTransaction(t *testing.T) {
	account := &BankAccount{
		value:               72,
		mux:                 sync.Mutex{},
		PendingTransactions: make(chan Transaction, 10), // Buffered channel to hold the transactions
		PostedTransactions:  make(chan Transaction, 10), // Buffered channel to hold the transactions
		transactionCount:    0,
		Pending:             []Transaction{},
	}

	account.addTransaction(30, Withdrawal)

	want := 1
	got := len(account.PendingTransactions)

	if got != want {
		t.Errorf("Channel does not contain value, expected %d, got %d", want, got)
	}

	want = 1
	got = len(account.Pending)
	if got != want {
		t.Errorf("Expected pending slice to hold a value of %d, got %d", want, got)
	}

	pendingTransaction := account.Pending[0]

	want = 30
	got = pendingTransaction.Value
	if got != want {
		t.Errorf("Transaction does not contain the right amount")
	}

	stringWant := "Withdrawal"
	stringGot := pendingTransaction.Type.String()

	if stringGot != stringWant {
		t.Errorf("Wrong transaction type added ")
	}
}

func BenchmarkMain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		main()
	}
}
