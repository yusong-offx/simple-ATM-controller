package atm

import (
	"errors"
	"log"
	"sync"

	"github.com/yusong-offx/myATM/bank"
)

type ATMID string

type ServiceType interface {
	Deposit(m bank.Money) error
	Withdrawal(m bank.Money) error
	Transfer(to bank.AccountNumber, m bank.Money) error
	GetHistory() []*bank.HistoryEle
	PINCheck(pin string) error
}

// You can make ATM info
// using ATM.ID entity as foreign key
// or just add this struct.
type ATM struct {
	ID           ATMID
	Cash         bank.Money
	StorageLimit bank.Money
	BasicDeposit bank.Money
	Service      ServiceType
	Logger       log.Logger
	Mutex        *sync.RWMutex
}

// You can add like : (limit uint64)
// const ATM_BASIC_DEPOSIT_1 = 10000
// const ATM_BASIC_DEPOSIT_2 = 20000
// ....
const ATM_BASIC_DEPOSIT = 10000

// Size of ATM cash storage
const ATM_STORAGE_LIMIT = 1000000

func NewATM() *ATM {
	return &ATM{
		ID:           ATMID(bank.GenerateUUID()),
		Cash:         ATM_BASIC_DEPOSIT,
		BasicDeposit: ATM_BASIC_DEPOSIT,
		StorageLimit: ATM_STORAGE_LIMIT,
		Mutex:        &sync.RWMutex{},
	}
}

// Get history from account as below format :
// "{time} {Action} {Action's Money} {Balance}"
func (a *ATM) GetHistory() []*bank.HistoryEle {
	if a.Service == nil {
		return []*bank.HistoryEle{}
	}
	return a.Service.GetHistory()
}

// Calculate amount of cash to make ATM_BASIC_DEPOSIT
// true: not enough / false: excess
func (a *ATM) CashBasicDeposit() (bank.Money, bool) {
	if a.BasicDeposit < a.Cash {
		return a.Cash - a.BasicDeposit, false
	}
	return a.BasicDeposit - a.Cash, true
}

// Read and save infomation of card or account, etc...
func (a *ATM) Read(object ServiceType, pin string) error {
	a.Mutex.Lock()
	a.Service = object

	if err := a.Service.PINCheck(pin); err != nil {
		a.Service = nil
		a.Mutex.Unlock()
		return err
	}
	return nil
}

func (a *ATM) Return() {
	a.Service = nil
	a.Mutex.Unlock()
}

func (a *ATM) Deposit(m bank.Money) error {
	if a.Service == nil {
		return errors.New("nothing to service")
	}
	if v, ok := bank.SafeAdd(a.Cash, m); !ok || v > a.StorageLimit {
		return errors.New("atm will overflow")
	}
	if err := a.Service.Deposit(m); err != nil {
		return err
	}
	a.Cash += m
	return nil
}

func (a *ATM) Withdrawal(m bank.Money) error {
	if a.Service == nil {
		return errors.New("nothing to service")
	}
	if a.Cash < m {
		return errors.New("not enough cash in atm")
	}
	if err := a.Service.Withdrawal(m); err != nil {
		return err
	}
	a.Cash -= m
	return nil
}

func (a *ATM) Transfer(to bank.AccountNumber, m bank.Money) error {
	if err := a.Service.Transfer(to, m); err != nil {
		return err
	}
	return nil
}
