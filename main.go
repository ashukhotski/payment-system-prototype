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
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --------------------------------------------------------
// Defining necessary enums such as account status and type as well as error message codes
type ErrorCode int8

const (
	AccountDoesNotExistError = iota
	AccountIsBlockedError
	InsufficientAccountBalanceError
	AccountTypeMismatchError
	AccountIbanMismatchError
	NegativeAmountError
	InvalidIbanError
	AccountCreationError
	AccountDetailsJsonError
	MoneyTransferJsonError
)

type LanguageCode int8

var locale LanguageCode

const (
	English = iota
	Russian
)

// Mapping error codes to messages
var errorCodesToMessagesMap map[ErrorCode](map[LanguageCode]string) = map[ErrorCode](map[LanguageCode]string){
	AccountDoesNotExistError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", AccountDoesNotExistError, "Requested account does not exist"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", AccountDoesNotExistError, "Запрашиваемый аккаунт не существует"),
	},
	AccountIsBlockedError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", AccountIsBlockedError, "Account is blocked"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", AccountIsBlockedError, "Аккаунт заблокирован"),
	},
	InsufficientAccountBalanceError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", InsufficientAccountBalanceError, "Insufficient account balance"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", InsufficientAccountBalanceError, "Недостаточно средств на балансе"),
	},
	AccountTypeMismatchError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", AccountTypeMismatchError, "Account has the wrong type"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", AccountTypeMismatchError, "Некорректный тип аккаунта"),
	},
	AccountIbanMismatchError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", AccountTypeMismatchError, "Account has the wrong IBAN"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", AccountTypeMismatchError, "Некорректный IBAN аккаунта"),
	},
	NegativeAmountError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", NegativeAmountError, "Amount cannot be negative"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", NegativeAmountError, "Сумма не может быть отрицательной"),
	},
	InvalidIbanError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", InvalidIbanError, "IBAN is not valid"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", InvalidIbanError, "IBAN не является валидным"),
	},
	AccountCreationError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", AccountCreationError, "Impossible to create account"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", AccountCreationError, "Невозможно создать аккаунт"),
	},
	AccountDetailsJsonError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", AccountDetailsJsonError, "Cannot represent accounts as JSON"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", AccountDetailsJsonError, "Невозможно преобразить аккаунты в JSON"),
	},
	MoneyTransferJsonError: {
		English: fmt.Sprintf("Error code: %d. Message: %s", MoneyTransferJsonError, "Cannot parse JSON"),
		Russian: fmt.Sprintf("Код ошибки: %d. Сообщение: %s", MoneyTransferJsonError, "Невозможно обработать JSON"),
	},
}

type AccountStatus int8

const (
	Active AccountStatus = iota
	Blocked
)

// Mapping account status codes to account status names considering locale
var accountStatusCodeToNameMap map[AccountStatus](map[LanguageCode]string) = map[AccountStatus](map[LanguageCode]string){
	Active: {
		English: "Active",
		Russian: "Активный",
	},
	Blocked: {
		English: "Blocked",
		Russian: "Заблокированный",
	},
}

type AccountType int8

const (
	Ordinary AccountType = iota
	MonetaryEmission
	MonetaryDestruction
)

// --------------------------------------------------------
// Defining account structure properties
type Account struct {
	Iban      string
	Status    AccountStatus
	Type      AccountType
	Balance   float64
	Fractions float64
	// can be augmented with account holder details
	// can be augmented with other properties such as the timestamp of last modification and so on
}

func round(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func roundAndExtractFractions(amount float64) (float64, float64) {
	var rounded float64 = math.Round(amount*100) / 100
	fractions := amount - rounded
	return rounded, fractions
}

func NewAccount(iban string, s AccountStatus, t AccountType, b float64) *Account {
	r, f := roundAndExtractFractions(b)
	return &Account{iban, s, t, r, f}
}

func (acc *Account) Block() {
	acc.Status = Blocked
}

func (acc *Account) Activate() {
	acc.Status = Active
}

func (acc *Account) Deduct(amount float64) {
	r, f := roundAndExtractFractions(amount)
	acc.Balance -= r
	acc.Fractions -= f

	acc.Balance = round(acc.Balance)
}

func (acc *Account) Add(amount float64) {
	r, f := roundAndExtractFractions(amount)
	acc.Balance += r
	acc.Fractions += f

	acc.Balance = round(acc.Balance)
}

// Helper functions to validate and generate IBAN
func IsValidIban(iban string) bool {
	// Stripping spaces since IBANs often contain them to separate characters in blocks of 4 for better readability
	iban = strings.Replace(iban, " ", "", -1)

	// Checking the length
	if len(iban) != 28 {
		return false
	}

	// Prepare an IBAN for mod-97 verification
	iban = iban[:4] + iban[4:]
	ibanConverted, err := ConvertIbanToNumericForm(iban)
	if err != nil {
		return false
	}

	// Perform mod-97 verification and return true or false depending on the results
	return Mod97(ibanConverted) == 1
}

// Converts an IBAN to its numeric string representation for mod-97 calculation.
func ConvertIbanToNumericForm(iban string) (string, error) {
	var numericBuilder strings.Builder
	for _, char := range iban {
		if char >= 'A' && char <= 'Z' {
			// Convert the letter to its corresponding number and append
			numericBuilder.WriteString(strconv.Itoa(int(char - 'A' + 10)))
			continue
		}
		if char >= '0' && char <= '9' {
			// Append the digit directly
			numericBuilder.WriteRune(char)
			continue
		}
		// Return an error for invalid characters
		return "", fmt.Errorf(errorCodesToMessagesMap[InvalidIbanError][locale])
	}
	return numericBuilder.String(), nil
}

// Mod97 calculates the modulo 97 of a large number represented as a string.
func Mod97(number string) int {
	var remainder int
	for i := 0; i < len(number); i++ {
		digit, _ := strconv.Atoi(string(number[i]))
		remainder = (remainder*10 + digit) % 97
	}
	return remainder
}

// Generates random Belarusian IBAN without ensuring its validity
func GenerateBelarusianIban() (string, error) {
	countryPrefix := "BY"

	const totalLength = 28
	const checkDigitsPlaceholder = "00" // Placeholder for check digits
	bbanLength := totalLength - 4       // Length of the Basic Bank Account Number (BBAN)

	// Generate a random BBAN with digits
	bban := GenerateRandomDigits(bbanLength)

	// Construct the IBAN with placeholder check digits
	iban := countryPrefix + checkDigitsPlaceholder + bban

	// Convert IBAN to numeric string for mod-97 calculation
	ibanNumeric, err := ConvertIbanToNumericForm(iban)
	if err != nil {
		return "", err
	}

	// Calculate check digits
	checkDigits := CalculateIbanCheckDigits(ibanNumeric)

	// Replace placeholder check digits in the IBAN
	iban = countryPrefix + checkDigits + bban

	return iban, nil
}

// Generates a random Belarusian IBAN that is valid
func GenerateValidBelarusianIban() (string, error) {
	var iban string = ""
	var err error = nil
	errCount := 0
	for !IsValidIban(iban) {
		// Breaking the loop if valid IBAN generation took too many tries
		// ideally the value to compare to errCount should be parsed from environmental configuration
		if errCount > 1000000 {
			return "", fmt.Errorf(errorCodesToMessagesMap[InvalidIbanError][locale])
		}
		// Attempting to generate a valid IBAN
		iban, err = GenerateBelarusianIban()
		if err != nil {
			errCount++
			continue
		}
		errCount++
	}
	return iban, nil
}

// Generates a string of random digits of a specified length.
func GenerateRandomDigits(length int) string {
	digits := make([]byte, length)
	for i := range digits {
		digits[i] = byte(rand.Intn(10) + '0')
	}
	return string(digits)
}

// Calculates the check digits for an IBAN given its numeric string representation.
func CalculateIbanCheckDigits(ibanNumeric string) string {
	// Perform mod-97 operation and subtract from 98 to get check digits
	checkValue := 98 - Mod97(ibanNumeric)
	checkDigits := fmt.Sprintf("%02d", checkValue)
	return checkDigits
}

// --------------------------------------------------------
// Defining implementation agnostic interface that contains methods to manipulate accounts
type AccountRepository interface {
	RetrieveEmissionAccountIban() (string, error)
	RetrieveDestructionAccountIban() (string, error)
	EmitMoney(amount float64) error
	DestructMoney(iban string, amount float64) error
	OpenAccount() (*Account, error)
	TransferMoney(sender, recipient string, amount float64) error
	TransferMoneyJson(jsonStr string) error
	RetrieveAllAccountsAsJson() (string, error)
	// Additional methods to manipulate the status of the account
	BlockAccount(iban string) error
	ActivateAccount(iban string) error
}

type AccountService struct {
	accountRepoImpl AccountRepository
}

func NewAccountService(r AccountRepository) *AccountService {
	return &AccountService{r}
}

func (s *AccountService) RetrieveEmissionAccountIban() (string, error) {
	return s.accountRepoImpl.RetrieveEmissionAccountIban()
}

func (s *AccountService) RetrieveDestructionAccountIban() (string, error) {
	return s.accountRepoImpl.RetrieveDestructionAccountIban()
}

func (s *AccountService) EmitMoney(amount float64) error {
	return s.accountRepoImpl.EmitMoney(amount)
}

func (s *AccountService) DestructMoney(iban string, amount float64) error {
	return s.accountRepoImpl.DestructMoney(iban, amount)
}

// Not passing account type assuming this method opens only ordinary accounts, not special accounts for monetary emmision and destruction
// Not passing account status assuming a newly opened account should be active immediately (this behavior can be change to comply with KYC)
// Not passing initial balance assuming it should only be topped up from the emission account by making a money transfer between accounts
func (s *AccountService) OpenAccount() (*Account, error) {
	return s.accountRepoImpl.OpenAccount()
}

func (s *AccountService) TransferMoney(sender, recipient string, amount float64) error {
	return s.accountRepoImpl.TransferMoney(sender, recipient, amount)
}

func (s *AccountService) TransferMoneyJson(jsonStr string) error {
	return s.accountRepoImpl.TransferMoneyJson(jsonStr)
}

func (s *AccountService) RetrieveAllAccountsAsJson() (string, error) {
	return s.accountRepoImpl.RetrieveAllAccountsAsJson()
}

func (s *AccountService) BlockAccount(iban string) error {
	return s.accountRepoImpl.BlockAccount(iban)
}

func (s *AccountService) ActivateAccount(iban string) error {
	return s.accountRepoImpl.ActivateAccount(iban)
}

// --------------------------------------------------------
// Defining in-memory implementation of account repository interface methods
// Explicitly declaring EmissionAccount and DestructionAccount properties for the ease of access (no need to iterate over a collection to get them)
type InMemoryAccountRepository struct {
	EmissionAccount    *Account
	DestructionAccount *Account
	Accounts           map[string]*Account // accounts decalred as map for speed and simplicity but array could be used instead
	Mutex              sync.Mutex
}

func NewInMemoryAccountRepository(eIban, dIban string) *InMemoryAccountRepository {
	eIban = strings.Replace(eIban, " ", "", -1)
	dIban = strings.Replace(dIban, " ", "", -1)
	emissionAcc := NewAccount(eIban, Active, MonetaryEmission, 0)
	destructionAcc := NewAccount(dIban, Active, MonetaryDestruction, 0)
	accounts := map[string]*Account{
		eIban: emissionAcc,
		dIban: destructionAcc,
	}
	return &InMemoryAccountRepository{emissionAcc, destructionAcc, accounts, sync.Mutex{}}
}

// Helper function to check if account with the given IBAN exists in the accounts map
func (r *InMemoryAccountRepository) accountExists(iban string) bool {
	if r.EmissionAccount != nil && r.EmissionAccount.Iban == iban {
		return true
	}
	if r.DestructionAccount != nil && r.DestructionAccount.Iban == iban {
		return true
	}
	_, exists := r.Accounts[iban]
	return exists
}

func (r *InMemoryAccountRepository) RetrieveEmissionAccountIban() (string, error) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	// Checking if emission account is set
	if r.EmissionAccount == nil {
		return "", fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Checking if account set as emission account is of the correct type
	if r.EmissionAccount.Type != MonetaryEmission {
		return "", fmt.Errorf(errorCodesToMessagesMap[AccountTypeMismatchError][locale])
	}
	return r.EmissionAccount.Iban, nil
}

func (r *InMemoryAccountRepository) RetrieveDestructionAccountIban() (string, error) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	// Checking if destruction account is set
	if r.DestructionAccount == nil {
		return "", fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Checking if account set as destruction account is of the correct type
	if r.DestructionAccount.Type != MonetaryDestruction {
		return "", fmt.Errorf(errorCodesToMessagesMap[AccountTypeMismatchError][locale])
	}
	return r.DestructionAccount.Iban, nil
}

func (r *InMemoryAccountRepository) EmitMoney(amount float64) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	// Checking if emission account is set
	if r.EmissionAccount == nil {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Checking if account set as emission account is of the correct type
	if r.EmissionAccount.Type != MonetaryEmission {
		return fmt.Errorf(errorCodesToMessagesMap[AccountTypeMismatchError][locale])
	}
	// Checking if the account is not blocked
	if r.EmissionAccount.Status == Blocked { // alternatively can be "if acc.Status != Active" depending on expected behavior
		return fmt.Errorf(errorCodesToMessagesMap[AccountIsBlockedError][locale])
	}
	// Checking if money amount to emit is not negative
	if amount < 0 {
		return fmt.Errorf(errorCodesToMessagesMap[NegativeAmountError][locale])
	}

	r.EmissionAccount.Add(amount)

	//TODO: send some kind of notification to message queue to be processed by transaction log microservice
	return nil
}

func (r *InMemoryAccountRepository) DestructMoney(iban string, amount float64) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	iban = strings.Replace(iban, " ", "", -1)

	// Checking if destruction account is set
	if r.DestructionAccount == nil {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Checking if account set as destruction account is of the correct type
	if r.DestructionAccount.Type != MonetaryDestruction {
		return fmt.Errorf(errorCodesToMessagesMap[AccountTypeMismatchError][locale])
	}
	// Checking if destruction account is not blocked
	if r.DestructionAccount.Status == Blocked { // alternatively can be "if acc.Status != Active" depending on expected behavior
		return fmt.Errorf(errorCodesToMessagesMap[AccountIsBlockedError][locale])
	}
	// Checking if money amount to deduct is not negative
	if amount < 0 {
		return fmt.Errorf(errorCodesToMessagesMap[NegativeAmountError][locale])
	}
	// Checking if account associated with the given IBAN exists
	if !r.accountExists(iban) {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	acc := r.Accounts[iban]
	// Ensuring that we indeed got the correct account object
	if acc.Iban != iban {
		return fmt.Errorf(errorCodesToMessagesMap[AccountIbanMismatchError][locale])
	}
	// Checking if the account is blocked (or is not active)
	if acc.Status == Blocked { // alternatively can be "if acc.Status != Active" depending on expected behavior
		return fmt.Errorf(errorCodesToMessagesMap[AccountIsBlockedError][locale])
	}
	// Checking if the account balance is sufficient to deduct the given amount
	if r, _ := roundAndExtractFractions(amount); acc.Balance < r {
		return fmt.Errorf(errorCodesToMessagesMap[InsufficientAccountBalanceError][locale])
	}

	acc.Deduct(amount)
	r.Accounts[acc.Iban] = acc
	r.DestructionAccount.Add(amount)

	//TODO: send some kind of notification to message queue to be processed by transaction log microservice
	return nil
}

func (r *InMemoryAccountRepository) OpenAccount() (*Account, error) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	iban := ""
	var err error = nil
	// Performing one or more attempts to generate a valid and unique Belarusian IBAN
	for iban == "" || (iban != "" && r.accountExists(iban)) {
		iban, err = GenerateValidBelarusianIban()
		if err != nil {
			return nil, fmt.Errorf(errorCodesToMessagesMap[AccountCreationError][locale])
		}
	}

	// Creating a new account and adding it to the account storage
	acc := NewAccount(iban, Active, Ordinary, 0)
	r.Accounts[iban] = acc
	return acc, nil
}

func (r *InMemoryAccountRepository) TransferMoney(sender, recipient string, amount float64) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	sender = strings.Replace(sender, " ", "", -1)
	recipient = strings.Replace(recipient, " ", "", -1)

	// Checking if sender account exists
	sAcc, sExists := r.Accounts[sender]
	if !sExists || sAcc == nil {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Ensuring that we indeed got the correct account object
	if sAcc.Iban != sender {
		return fmt.Errorf(errorCodesToMessagesMap[AccountIbanMismatchError][locale])
	}
	// Checking if sender account is not blocked
	if sAcc.Status == Blocked { // alternatively can be "if acc.Status != Active" depending on expected behavior
		return fmt.Errorf(errorCodesToMessagesMap[AccountIsBlockedError][locale])
	}
	// Checking if money amount to transfer is not negative
	if amount < 0 {
		return fmt.Errorf(errorCodesToMessagesMap[NegativeAmountError][locale])
	}
	//Checking if sender has sufficient balance to transfer the amount to recipient
	if r, _ := roundAndExtractFractions(amount); sAcc.Balance < r {
		return fmt.Errorf(errorCodesToMessagesMap[InsufficientAccountBalanceError][locale])
	}
	// Checking if recipient account exists
	rAcc, rExists := r.Accounts[recipient]
	if !rExists {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Ensuring that we indeed got the correct account object
	if rAcc.Iban != recipient {
		return fmt.Errorf(errorCodesToMessagesMap[AccountIbanMismatchError][locale])
	}
	// Checking if recipient account is not blocked
	if rAcc.Status == Blocked {
		return fmt.Errorf(errorCodesToMessagesMap[AccountIsBlockedError][locale])
	}
	// TODO: prohibit transfer for certain account types if it makes sense (i.e., cannot send from ordinary account to monetary emission account)

	sAcc.Deduct(amount)
	r.Accounts[sender] = sAcc
	rAcc.Add(amount)
	r.Accounts[recipient] = rAcc

	//TODO: send some kind of notification to message queue to be processed by transaction log microservice
	return nil
}

func (r *InMemoryAccountRepository) TransferMoneyJson(jsonStr string) error {
	type moneyTransferReq struct {
		Sender    string  `json:"sender"`
		Recipient string  `json:"recipient"`
		Amount    float64 `json:"amount"`
	}
	var req moneyTransferReq
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return fmt.Errorf(errorCodesToMessagesMap[MoneyTransferJsonError][locale])
	}
	return r.TransferMoney(req.Sender, req.Recipient, req.Amount)
}

func (r *InMemoryAccountRepository) RetrieveAllAccountsAsJson() (string, error) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	type accountDetails struct {
		Iban      string  `json:"iban"`
		Balance   float64 `json:"balance"`
		Fractions float64 `json:"fractions"`
		Status    string  `json:"status"`
	}
	allAccountDetails := []accountDetails{}
	if r.EmissionAccount != nil {
		allAccountDetails = append(allAccountDetails, accountDetails{r.EmissionAccount.Iban, r.EmissionAccount.Balance, r.EmissionAccount.Fractions, accountStatusCodeToNameMap[r.EmissionAccount.Status][locale]})
	}
	if r.DestructionAccount != nil {
		allAccountDetails = append(allAccountDetails, accountDetails{r.DestructionAccount.Iban, r.DestructionAccount.Balance, r.DestructionAccount.Fractions, accountStatusCodeToNameMap[r.DestructionAccount.Status][locale]})
	}
	for _, acc := range r.Accounts {
		if acc != r.EmissionAccount && acc != r.DestructionAccount {
			allAccountDetails = append(allAccountDetails, accountDetails{acc.Iban, acc.Balance, acc.Fractions, accountStatusCodeToNameMap[acc.Status][locale]})
		}
	}
	output, err := json.Marshal(allAccountDetails)
	if err != nil {
		return "", fmt.Errorf(errorCodesToMessagesMap[AccountDetailsJsonError][locale])
	}
	return string(output), nil
}

func (r *InMemoryAccountRepository) BlockAccount(iban string) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	iban = strings.Replace(iban, " ", "", -1)

	// Checking if account associated with the given IBAN exists
	if !r.accountExists(iban) {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	acc := r.Accounts[iban]
	// Ensuring that account object is not nil
	if acc == nil {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Ensuring that we indeed got the correct account object
	if acc.Iban != iban {
		return fmt.Errorf(errorCodesToMessagesMap[AccountIbanMismatchError][locale])
	}

	acc.Block()
	r.Accounts[acc.Iban] = acc
	return nil
}

func (r *InMemoryAccountRepository) ActivateAccount(iban string) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	iban = strings.Replace(iban, " ", "", -1)

	// Checking if account associated with the given IBAN exists
	if !r.accountExists(iban) {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	acc := r.Accounts[iban]
	// Ensuring that account object is not nil
	if acc == nil {
		return fmt.Errorf(errorCodesToMessagesMap[AccountDoesNotExistError][locale])
	}
	// Ensuring that we indeed got the correct account object
	if acc.Iban != iban {
		return fmt.Errorf(errorCodesToMessagesMap[AccountIbanMismatchError][locale])
	}

	acc.Activate()
	r.Accounts[acc.Iban] = acc
	return nil
}

// --------------------------------------------------------
// Initializing the app and assigning values to certain parameters
// Ideally, those should be parsed from the environment configuration or vault
func init() {
	rand.Seed(time.Now().UnixNano())
	locale = English
}

func main() {
	inMemRepoImpl := NewInMemoryAccountRepository("BY84 ALFA 1000 0000 0000 0000 0000", "BY84 ALFA 1000 0000 0000 0000 0001")
	service := NewAccountService(inMemRepoImpl)

	wg := sync.WaitGroup{}

	// Get IBAN of emission account
	testGettingEmissionIBAN(service)

	// Get IBAN of destruction account
	testGettingDestructionIBAN(service)

	// Attempt to open a new ordinary account and topping up the balance (failure)
	testAccountOpeningAndTopupFailure(service)

	// Attempt to open a new ordinary account with zero balance (success)
	testZeroBalanceAccountOpening(service)

	// Open multiple ordinary accounts and topping up the balances in parallel
	const n int = 20
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testAccountOpeningAndTopupSuccess(service)
		}()
	}
	wg.Wait()

	// Attempt to destruct money (failure)
	testMoneyDestructionFailure(service)

	// Attempt to emit money (success)
	testMoneyEmissionSuccess(service)

	// Attempt to destruct money (success)
	testMoneyDestructionSuccess(service)

	// Attempt to transfer money between accounts (success)
	testSuccessfulMoneyTransfer(service)

	// Attempt to transfer money between accounts (failure)
	testFailedMoneyTransfer(service)

	// Testing concurrency by performing M money transfers in parallel
	const m int = 100
	for i := 0; i < m; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMoneyTransferViaJson(service)
		}()
	}
	wg.Wait()

	// Print all accounts details
	testAllAccountDetailsPrinting(service)
}

// Get IBAN of emission account
func testGettingEmissionIBAN(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 1: getting emission account IBAN\n")
	iban, err := service.RetrieveEmissionAccountIban()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN: %s", iban))
	fmt.Println(builder.String())
}

// Get IBAN of destruction account
func testGettingDestructionIBAN(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 2: getting destruction account IBAN\n")
	iban, err := service.RetrieveDestructionAccountIban()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN: %s", iban))
	fmt.Println(builder.String())
}

// Open a new ordinary account and topping up the balance (failure)
func testAccountOpeningAndTopupFailure(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 3: failing to open a new account and top up its balance\n")
	acc, err := service.OpenAccount()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	err = service.TransferMoney("BY84 ALFA 1000 0000 0000 0000 0000", acc.Iban, -23.48)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
	}
}

// Open a new ordinary account and topping up the balance (success)
func testAccountOpeningAndTopupSuccess(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 4: presumably successfully opening a new account and topping up its balance\n")
	acc, err := service.OpenAccount()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	var amount float64 = rand.Float64() * float64(rand.Intn(1000))
	err = service.EmitMoney(amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	err = service.TransferMoney("BY84 ALFA 1000 0000 0000 0000 0000", acc.Iban, amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN %s: %.2f", acc.Iban, round(amount)))
	fmt.Println(builder.String())
}

// Open a new ordinary account with zero balance (success)
func testZeroBalanceAccountOpening(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 5: presumably successfully opening an account with zero balance\n")
	acc, err := service.OpenAccount()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("IBAN %s: %.2f", acc.Iban, acc.Balance))
	fmt.Println(builder.String())
}

// Destruct money (failure)
func testMoneyDestructionFailure(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 6: failing to destruct money\n")
	err := service.DestructMoney("BY84 ALFA 1000 0000 0000 0000 0000", -10000)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
	}
}

// Emit money (success)
func testMoneyEmissionSuccess(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 7: presumably successfully emitting money\n")
	var amount float64 = 250
	err := service.EmitMoney(amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("Money emitted: %.2f\n", round(amount)))
	fmt.Println(builder.String())
}

// Destruct money (success)
func testMoneyDestructionSuccess(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 8: presumably successfully destructing money\n")
	var amount float64 = 10
	iban := "BY84 ALFA 1000 0000 0000 0000 0000"
	err := service.DestructMoney(iban, amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("Money destructed to %s: %.2f\n", iban, round(amount)))
	fmt.Println(builder.String())
}

// Print all accounts details
func testAllAccountDetailsPrinting(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 9: printing IBAN, balance and status of all existing accounts including special and ordinary\n")
	res, err := service.RetrieveAllAccountsAsJson()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, res)
	fmt.Println(builder.String())
}

// Transfer money between accounts (success)
func testSuccessfulMoneyTransfer(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 10: presumably successfully transferring money between accounts\n")
	sender := "BY84 ALFA 1000 0000 0000 0000 0000"
	recipient := "BY84 ALFA 1000 0000 0000 0000 0001"
	var amount float64 = 50
	err := service.TransferMoney(sender, recipient, amount)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	fmt.Fprintf(&builder, fmt.Sprintf("Money transfer from %s to %s: %.2f\n", sender, recipient, round(amount)))
	fmt.Println(builder.String())
}

// Transfer money between accounts (failure)
func testFailedMoneyTransfer(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 11: failing to transfer money between accounts\n")
	// Blocking an account to fail the subsequent money transfer attempt
	err := service.BlockAccount("BY84 ALFA 1000 0000 0000 0000 0000")
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
		return
	}
	err = service.TransferMoney("BY84 ALFA 1000 0000 0000 0000 0000", "BY84 ALFA 1000 0000 0000 0000 0001", 50)
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
	}
	// Activating account again to remove the block set earlier, so that future operations won't be affected by this use case
	err = service.ActivateAccount("BY84 ALFA 1000 0000 0000 0000 0000")
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
	}
	fmt.Println(builder.String())
}

// Picking two random accounts and transferring money between them via JSON request
func testMoneyTransferViaJson(service *AccountService) {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Use Case 12: picking two random accounts and transferring money between them\n")
	str, err := service.RetrieveAllAccountsAsJson()
	if err != nil {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", err))
		fmt.Println(builder.String())
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
		fmt.Println(builder.String())
		return
	}

	// Excluding special accounts from consideration and shuffling remaining ordinary accounts
	if len(accounts) < 4 {
		fmt.Fprintf(&builder, fmt.Sprintf("Error: %v\n", "Not enough accounts to execute use case 12"))
		fmt.Println(builder.String())
		return
	}
	accounts = accounts[2:]
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
	fmt.Fprintf(&builder, fmt.Sprintf("Money transfer from %s to %s: %.2f\n", mt.Sender, mt.Recipient, round(mt.Amount)))
	fmt.Println(builder.String())
}
