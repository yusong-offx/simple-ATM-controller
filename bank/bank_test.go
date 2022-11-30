package bank_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/yusong-offx/myATM/bank"
)

const (
	PWD    = "0000"
	CREATE = 1
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	testclient  *bank.Client
	testaccount *bank.Account
	testcard    *bank.Card
	err         error
)

func randStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Create test
func TestCreateClient(t *testing.T) {
	testclient, err = bank.NewClient(&bank.Client{SSN: bank.SSN(bank.GenerateUUID()),
		Name: bank.Name{
			Frist: randStringRunes(7),
			Last:  randStringRunes(5),
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCreateAccount(t *testing.T) {
	testaccount, err = testclient.NewAccount(PWD)
	if err != nil {
		t.Error(err)
	}
}

// One acccount can bind maximum (default 3)
func TestCreateCard(t *testing.T) {
	// Create test
	testcard, err = testaccount.NewCard(PWD)
	if err != nil {
		t.Error(err)
	}
	// Add 2 cards,
	// It could make same card number error by random uuid
	for i := 0; i < 2; i++ {
		_, err := testaccount.NewCard(PWD)
		if err != nil {
			t.Error(err)
		}
	}
	// Make error
	_, err = testaccount.NewCard(PWD)
	if err == nil {
		t.Error("exception error")
	}
}

func TestAccountDeposit(t *testing.T) {
	before := testaccount.Balance
	if err := testaccount.Deposit(100); err != nil {
		t.Error(err)
	}
	if before+100 != testaccount.Balance {
		t.Error("deposit error")
	}
	// Overflow test
	var small bank.Money = 0
	if err := testaccount.Deposit(^small); err == nil {
		t.Error(err)
	}
}

func TestAccountWithdrawal(t *testing.T) {
	before := testaccount.Balance
	if err := testaccount.Withdrawal(100); err != nil {
		t.Error(err)
	}
	if before-100 != testaccount.Balance {
		t.Error("withdrawal error")
	}
	// Not enough balance
	if err := testaccount.Withdrawal(100); err == nil {
		t.Error(err)
	}
}

func TestCardDeposit(t *testing.T) {
	before := testaccount.Balance
	if err := testaccount.Deposit(100); err != nil {
		t.Error(err)
	}
	if before+100 != testaccount.Balance {
		t.Error("deposit error")
	}
	// Overflow test
	var small bank.Money = 0
	if err := testaccount.Deposit(^small); err == nil {
		t.Error(err)
	}
}

func TestCardWithdrawal(t *testing.T) {
	before := testaccount.Balance
	if err := testaccount.Withdrawal(100); err != nil {
		t.Error(err)
	}
	if before-100 != testaccount.Balance {
		t.Error("withdrawal error")
	}
	// Not enough balance
	if err := testaccount.Withdrawal(100); err == nil {
		t.Error(err)
	}
}

func TestAccountTransfer(t *testing.T) {
	// Ignore create error
	newAccount, _ := testclient.NewAccount(PWD)
	if err := testaccount.Deposit(100); err != nil {
		t.Error(err)
	}
	before := testaccount.Balance
	if err := testaccount.Transfer(newAccount.Number, 100); err != nil {
		t.Error(err)
	}
	if !(before-100 == testaccount.Balance && newAccount.Balance == bank.Money(100)) {
		t.Error("transfer error")
	}
}

func TestCardTransfer(t *testing.T) {
	// Ignore error
	account, _ := testclient.NewAccount(PWD)
	card, _ := account.NewCard(PWD)
	card.Deposit(100)
	before := card.Balance()
	if err := card.Transfer(bank.AccountFromCardNumber[testcard.Number].Number, 100); err != nil {
		t.Error(err)
	}
	if !(before-100 == card.Balance() && testcard.Balance() == bank.Money(100)) {
		t.Error("transfer error")
	}
}
