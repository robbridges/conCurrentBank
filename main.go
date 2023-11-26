package main

func main() {
	transactions := []struct {
		value           int
		transactionType TransactionType
	}{
		{100, Deposit},
		{50, Withdrawal},
		{20, Deposit},
		{30, Deposit},
		{5, Withdrawal},
	}

	startBank(transactions)
}
