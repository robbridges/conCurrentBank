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
		PendingTransactions: make(chan Transaction, 10),
		PostedTransactions:  make(chan Transaction, 10),
		transactionCount:    0,
		Pending:             []Transaction{},
	}

	t.Run("Happy path", func(t *testing.T) {
		err := account.addTransaction(30, Withdrawal)
		if err != nil {
			t.Errorf("Recieved unexpected error when adding transaction")
		}

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
	})
	t.Run("Sad path, channel full", func(t *testing.T) {
		account.PendingTransactions = make(chan Transaction, 1)
		err := account.addTransaction(20, Deposit)
		if err != nil {
			t.Errorf("Unexpected error when adding first transaction")
		}
		got := len(account.Pending)
		want := 1
		if got != want {
			t.Errorf("Successfully adding a transaction should have been appended to pending slice")
		}

		err = account.addTransaction(5, Deposit)
		if err == nil {
			t.Errorf("Expected error due to channel being full")
		}

		// the slice should still be on length one, the failed transaction should not have been added
		got = len(account.Pending)
		want = 1
		if got != want {
			t.Errorf("The failed transaction was added to the pending slice")
		}

		stringGot := err.Error()
		stringWant := "Unable to add transaction"
		if stringGot != stringWant {
			t.Errorf("Wrong error message returned, error:%s, wanted:%s", stringGot, stringWant)
		}
	})
}

func TestProcessTransactions(t *testing.T) {
	account := &BankAccount{
		value:               72,
		mux:                 sync.Mutex{},
		PendingTransactions: make(chan Transaction, 10),
		PostedTransactions:  make(chan Transaction, 10),
		transactionCount:    0,
		Pending:             make([]Transaction, 0),
		Posted:              make([]Transaction, 0),
	}
	t.Run("Happy path withdrawal", func(t *testing.T) {
		err := account.addTransaction(30, Withdrawal)
		if err != nil {
			t.Errorf("Expected no error adding to transaction")
		}
		// need to close the pendingtransactions channel when we're doing adding transactions
		close(account.PendingTransactions)
		err = account.processTransactions()
		if err != nil {
			t.Errorf("Unexpected error processing transaction")
		}
		want := 0
		got := len(account.Pending)
		if want != got {
			t.Errorf("Expected the pending transactions to be cleared during processing")
		}

		want = 1
		got = len(account.PostedTransactions)
		if want != got {
			t.Errorf("Expected the posted transaction to be in the slice")
		}

		want = 42
		got = account.getBalance()

		if want != got {
			t.Errorf("The bank account balance should have been updated from the withdrawal")
		}
	})
	t.Run("Happy Path Deposit", func(t *testing.T) {
		err := account.addTransaction(55, Deposit)
		if err != nil {
			t.Errorf("Expected no error adding to transaction")
		}
		// need to close the pendingtransactions channel when we're doing adding transactions
		close(account.PendingTransactions)
		err = account.processTransactions()
		if err != nil {
			t.Errorf("Unexpected error processing transaction")
		}
		want := 0
		got := len(account.Pending)
		if want != got {
			t.Errorf("Expected the pending transactions to be cleared during processing")
		}

		want = 1
		got = len(account.PostedTransactions)
		if want != got {
			t.Errorf("Expected the posted transaction to be in the slice")
		}

		want = 127
		got = account.getBalance()

		if want != got {
			t.Errorf("The bank account balance should have been updated from the Deposit: got %d, expected: %d", got, want)
		}
	})
	t.Run("Sad Path insufficent funds", func(t *testing.T) {
		err := account.addTransaction(200, Withdrawal)
		if err != nil {
			t.Errorf("Unexpected error processing withdrawal")
		}
		close(account.PendingTransactions)
		err = account.processTransactions()
		if err == nil {
			t.Errorf("expected error withdrawing amount")
		}
		// transaction should not have processed and bank account should have been the same
		if account.getBalance() != 72 {
			t.Errorf("The value should have never been deducted")
		}
	})
}

func BenchmarkMain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		main()
	}
}
