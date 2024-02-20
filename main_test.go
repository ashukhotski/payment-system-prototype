// Author: Alexandr Shukhotskiy
// Task:
// - implement a prototype of payments processing system
// - ensure an account entity contains IBAN, balance and status attributes
// - ensure there are two special accounts (one for money emission and one for money destruction) besides ordinary accounts
// - ensure the system supports the following business functions:
// -- ability to get IBAN of the money emission account
// -- ability to get IBAN of the money destruction account
// -- ability to emit money to the emission account
// -- ability to destruct money from a given IBAN to the destruction account
// -- ability to open a new account
// -- ability to transfer money between any two accounts (either by passing variable parameters or json string)
// -- ability to list IBAN, remaining balance and status of all existing accounts (emission, destruction, ordinary) in JSON format
// System behavior can be tested by hardcoding several use case scenarios and executing them within the main function
// Notes:
// - I expanded the prototype with a few methods not mentioned in the original requirements to make it more complete:
// -- methods to block and activate the account
// Areas for improvement:
// - refactor the code and split it into packages and files
// - introduce service layer for external communication (http, grcp or tcp)
// - add more methods to manipulate account repository
// - introduce the queue (can be done via Go channels) and push messages there upon certain methods execution
// - leverage blockchain or linked list data structure to implement financial transactions log, possibly integrate it with the queue
// - implement better unit tests (a good task to delegate to junior and middle level developers)
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// Get IBAN of emission account
func TestGettingEmissionIBAN(t *testing.T) {
	emission := "BY84ALFA10000000000000000000"
	destruction := "BY84ALFA10000000000000000001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 1: getting emission account IBAN\n")
	iban, err := service.RetrieveEmissionAccountIban()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	if iban != emission {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v (%s != %s)\n", "Emission IBAN mismatch error", iban, emission))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN: %s", iban))
	fmt.Println(builder.String())
}

// Get IBAN of destruction account
func TestGettingDestructionIBAN(t *testing.T) {
	emission := "BY84ALFA10000000000000000000"
	destruction := "BY84ALFA10000000000000000001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 2: getting destruction account IBAN\n")
	iban, err := service.RetrieveDestructionAccountIban()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	if iban != destruction {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", "Destruction IBAN mismatch error"))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN: %s", iban))
	fmt.Println(builder.String())
}

// Open a new ordinary account and topping up the balance (failure)
func TestAccountOpeningAndTopupFailure(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 3: failing to open a new account and top up its balance\n")
	acc, err := service.OpenAccount()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	err = service.TransferMoney(emission, acc.Iban, -23.48)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	t.Errorf("Account opeing failed to fail")
}

// Open a new ordinary account and topping up the balance (success)
func TestAccountOpeningAndTopupSuccess(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 4: presumably successfully opening a new account and topping up its balance\n")
	acc, err := service.OpenAccount()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	var amount float64 = rand.Float64() * float64(rand.Intn(1000))
	err = service.EmitMoney(amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	err = service.TransferMoney(emission, acc.Iban, amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN %s: %f", acc.Iban, amount))
	fmt.Println(builder.String())
}

// Open a new ordinary account with zero balance (success)
func TestZeroBalanceAccountOpening(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 5: presumably successfully opening an account with zero balance\n")
	acc, err := service.OpenAccount()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN %s: %f", acc.Iban, acc.Balance))
	fmt.Println(builder.String())
}

// Destruct money (failure)
func TestMoneyDestructionFailure(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 6: failing to destruct money\n")
	err := service.DestructMoney(emission, -10000)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	t.Errorf("Money destruction failed to fail")
}

// Emit money (success)
func TestMoneyEmissionSuccess(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 7: presumably successfully emitting money\n")
	var amount float64 = 250
	err := service.EmitMoney(amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("Money emitted: %f\n", amount))
	fmt.Println(builder.String())
}

// Destruct money (success)
func TestMoneyDestructionSuccess(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 8: presumably successfully destructing money\n")
	var amount float64 = 250
	err := service.EmitMoney(amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	iban := "BY84 ALFA 1000 0000 0000 0000 0000"
	err = service.DestructMoney(emission, amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("Money destructed to %s: %f\n", iban, amount))
	fmt.Println(builder.String())
}

// Print all accounts details
func TestAllAccountDetailsPrinting(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 9: printing IBAN, balance and status of all existing accounts including special and ordinary\n")
	res, err := service.RetrieveAllAccountsAsJson()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, res)
	fmt.Println(builder.String())
}

// Transfer money between accounts (success)
func TestSuccessfulMoneyTransfer(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 10: presumably successfully transferring money between accounts\n")
	var amount float64 = 250
	err := service.EmitMoney(amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	sender := "BY84 ALFA 1000 0000 0000 0000 0000"
	recipient := "BY84 ALFA 1000 0000 0000 0000 0001"
	amount = 50
	err = service.TransferMoney(sender, recipient, amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("Money transfer from %s to %s: %f\n", sender, recipient, amount))
	fmt.Println(builder.String())
}

// Transfer money between accounts (failure)
func TestFailedMoneyTransfer(t *testing.T) {
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 11: failing to transfer money between accounts\n")
	// Blocking an account to fail the subsequent money transfer attempt
	err := service.BlockAccount(emission)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	err = service.TransferMoney(emission, destruction, 50)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
	}
	// Activating account again to remove the block set earlier, so that future operations won't be affected by this use case
	err2 := service.ActivateAccount(emission)
	if err2 != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err2))
	}
	fmt.Println(builder.String())
	if err != nil || err2 != nil {
		fmt.Println(builder.String())
		return
	}
	t.Errorf("Money transfer failed to fail")
}

// Picking two random accounts and transferring money between them via JSON request
func TestMoneyTransferViaJson(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	emission := "BY84 ALFA 1000 0000 0000 0000 0000"
	destruction := "BY84 ALFA 1000 0000 0000 0000 0001"
	inMemImpl := NewInMemoryAccountRepository(emission, destruction)
	service := NewAccountService(inMemImpl)

	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 12: picking two random accounts and transferring money between them\n")

	wg := sync.WaitGroup{}
	const n int = 50
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			acc, err := service.OpenAccount()
			if err != nil {
				fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
				t.Errorf(builder.String())
				return
			}
			var amount float64 = rand.Float64() * float64(rand.Intn(1000))
			err = service.EmitMoney(amount)
			if err != nil {
				fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
				t.Errorf(builder.String())
				return
			}
			err = service.TransferMoney(emission, acc.Iban, amount)
			if err != nil {
				fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
				t.Errorf(builder.String())
				return
			}
		}()
	}
	wg.Wait()

	str, err := service.RetrieveAllAccountsAsJson()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}
	type accountDetails struct {
		Iban    string  `json:"iban"`
		Balance float64 `json:"balance"`
		Status  string  `json:"status"`
	}
	var accounts []accountDetails
	if err := json.Unmarshal([]byte(str), &accounts); err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		t.Errorf(builder.String())
		return
	}

	// Excluding special accounts from consideration and shuffling remaining ordinary accounts
	if len(accounts) < 4 {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", "Not enough accounts to execute use case 12"))
		t.Errorf(builder.String())
		return
	}
	accounts = accounts[2:]

	const m int = 1000
	for i := 0; i < m; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rand.Shuffle(len(accounts), func(i, j int) { accounts[i], accounts[j] = accounts[j], accounts[i] })

			type moneyTransfer struct {
				Sender    string  `json:"sender"`
				Recipient string  `json:"recipient"`
				Amount    float64 `json:"amount"`
			}

			mt := moneyTransfer{accounts[0].Iban, accounts[1].Iban, rand.Float64() + float64(rand.Intn(100))}

			jsonStr, err := json.Marshal(mt)
			if err != nil {
				fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", "Unable to execute use case 12 due to JSON related error"))
				fmt.Println(builder.String())
				return
			}
			fmt.Fprintf(&builder, fmt.Sprintf("JSON: %s\n", string(jsonStr)))

			err = service.TransferMoneyJson(string(jsonStr))
			if err != nil {
				fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
				fmt.Println(builder.String())
				return
			}
			fmt.Fprintf(&builder, fmt.Sprintf("Money transfer from %s to %s: %f\n", mt.Sender, mt.Recipient, mt.Amount))
		}()
	}
	wg.Wait()

	fmt.Println(builder.String())
}
