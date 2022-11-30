package bank

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type WrapClient struct {
	Mutex *sync.RWMutex
	Value *Client
}

type WrapAccount struct {
	Mutex *sync.RWMutex
	Value *Account
}

type WrapCard struct {
	Mutex *sync.RWMutex
	Value *Card
}

// You can use enum or
// more efficient struct in real
type HistoryEle struct {
	T          time.Time
	Action     string
	Trader     AccountNumber
	Amount     Money
	AccBalance Money
}

// This is test source.
var (
	SyncClientFromSSN       = map[SSN]WrapClient{}
	SyncAccountFromNumber   = map[AccountNumber]WrapAccount{}
	SyncCardFromNumber      = map[CardNumber]WrapCard{}
	ClientFromAccountNumber = map[AccountNumber]*Client{}
	AccountFromCardNumber   = map[CardNumber]*Account{}
	CardFromAccountNumber   = map[AccountNumber]map[CardNumber]*Card{}
)

// Using uuid for
// ssn, account/card number and pin.
func GenerateUUID() string {
	return uuid.Must(uuid.NewRandom()).String()
}

func InsertClientDB(c *Client) {
	SyncClientFromSSN[c.SSN] = WrapClient{
		Mutex: &sync.RWMutex{},
		Value: c,
	}
}

func InsertCardDB(a *Account, c *Card) {
	SyncCardFromNumber[c.Number] = WrapCard{
		Mutex: &sync.RWMutex{},
		Value: c,
	}
	if _, ok := CardFromAccountNumber[a.Number]; !ok {
		CardFromAccountNumber[a.Number] = map[CardNumber]*Card{}
	}
	CardFromAccountNumber[a.Number][c.Number] = c
	AccountFromCardNumber[c.Number] = a
}

func InsertAccountDB(c *Client, a *Account) {
	SyncAccountFromNumber[a.Number] = WrapAccount{
		Mutex: &sync.RWMutex{},
		Value: a,
	}
	ClientFromAccountNumber[a.Number] = c
}
