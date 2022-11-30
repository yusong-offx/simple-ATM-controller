package bank

import (
	"errors"
	"time"
)

type CardNumber string
type CVC string
type CardPIN string

// Default validthru period (month)
// When you re-issue card, you have to
// check account validthru
const CARD_VALIDTHRU_PERIOD = 36

type Card struct {
	Number    CardNumber
	CVC       CVC
	PIN       CardPIN
	ValidThru time.Time
}

func (c *Card) Deposit(m Money) error {
	cardMutex := SyncCardFromNumber[c.Number].Mutex
	bindAccount := AccountFromCardNumber[c.Number]
	cardMutex.Lock()
	err := bindAccount.Deposit(m)
	cardMutex.Unlock()
	return err
}

func (c *Card) Withdrawal(m Money) error {
	cardMutex := SyncCardFromNumber[c.Number].Mutex
	bindAccount := AccountFromCardNumber[c.Number]
	cardMutex.Lock()
	err := bindAccount.Withdrawal(m)
	cardMutex.Unlock()
	return err
}

func (c *Card) Transfer(to AccountNumber, m Money) error {
	cardMutex := SyncCardFromNumber[c.Number].Mutex
	bindAccount := AccountFromCardNumber[c.Number]
	cardMutex.Lock()
	err := bindAccount.Transfer(to, m)
	cardMutex.Unlock()
	return err
}

// Card history show bind account history
func (c *Card) GetHistory() []*HistoryEle {
	bindAccount := AccountFromCardNumber[c.Number]
	return bindAccount.GetHistory()
}

func (c *Card) Balance() Money {
	bindaccount := AccountFromCardNumber[c.Number]
	return bindaccount.Balance
}

func (c *Card) PINCheck(pin string) error {
	if CardPIN(pin) != c.PIN {
		return errors.New("wrong PIN")
	}
	return nil
}
