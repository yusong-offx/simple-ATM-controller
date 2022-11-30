package atm_test

import (
	"testing"

	"github.com/yusong-offx/myATM/atm"
	"github.com/yusong-offx/myATM/bank"
)

type Person struct {
	Client  *bank.Client
	Account *bank.Account
	Card    *bank.Card
}

const PWD = "123"

var (
	person1 Person
	person2 Person
	myatm   *atm.ATM
)

func setTest() {
	myatm = atm.NewATM()
	person1 = Person{}
	person1.Client, _ = bank.NewClient(&bank.Client{SSN: bank.SSN(bank.GenerateUUID())})

	person1.Account, _ = person1.Client.NewAccount(PWD)
	person1.Card, _ = person1.Account.NewCard(PWD)

	person2 = Person{}
	person2.Client, _ = bank.NewClient(&bank.Client{SSN: bank.SSN(bank.GenerateUUID())})

	person2.Account, _ = person2.Client.NewAccount(PWD)
	person2.Card, _ = person2.Account.NewCard(PWD)
}

func TestATMCashBasicDeposit(t *testing.T) {
	setTest()
	myatm.Mutex.Lock()
	// For test temporarily subtract
	myatm.Cash -= 100
	v, flag := myatm.CashBasicDeposit()
	if !(flag && v == 100) {
		t.Error("cash basic error")
	}
	myatm.Cash += v
	myatm.Mutex.Unlock()
}

// Read and Return
func TestRead(t *testing.T) {
	// Mutex lock
	if err := myatm.Read(person1.Account, PWD); err != nil {
		t.Error(err)
	}
	// Lock check
	if myatm.Mutex.TryLock() {
		t.Error("mutex error")
	}
	// Unlock
	myatm.Return()
	// Unlock check
	if !myatm.Mutex.TryLock() {
		t.Error("mutex error")
	}
	myatm.Mutex.Unlock()
}

func TestATMDeposit(t *testing.T) {
	var err error
	if err = myatm.Read(person1.Account, "000"); err == nil {
		t.Error("PIN lock error")
	}
	if err = myatm.Read(person1.Card, "000"); err == nil {
		t.Error("PIN lock error")
	}
	if err = myatm.Read(person1.Account, PWD); err != nil {
		t.Error(err)
	}
	// Deposit test
	before_atm, before_account := myatm.Cash, person1.Account.Balance
	if err = myatm.Deposit(100); err != nil {
		t.Error(err)
	}
	if !(before_atm+100 == myatm.Cash && before_account+100 == person1.Account.Balance) {
		t.Error("atm deposit err")
	}
	// ATM limit test
	if err = myatm.Deposit(myatm.StorageLimit); err == nil {
		t.Error("atm over storage limit")
	}
	myatm.Return()
}

func TestWithdrawal(t *testing.T) {
	var err error
	if err = myatm.Read(person1.Account, PWD); err != nil {
		t.Error(err)
	}
	// Withdrawal test
	before_atm, before_account := myatm.Cash, person1.Account.Balance
	if err = myatm.Withdrawal(100); err != nil {
		t.Error(err)
	}
	if !(before_atm-100 == myatm.Cash && before_account-100 == person1.Account.Balance) {
		t.Error("atm Withdrawal err")
	}
	// ATM not enough cash
	if err = myatm.Withdrawal(myatm.Cash + 1); err == nil {
		t.Error("atm not enough cash to pay")
	}
	myatm.Return()
}
