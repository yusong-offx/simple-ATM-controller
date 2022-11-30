package bank

import (
	"errors"
	"fmt"
	"time"
)

type AccountNumber string
type AccountPIN string
type Money uint64

type Account struct {
	Number    AccountNumber
	PIN       AccountPIN
	ValidThru time.Time
	Balance   Money
	History   []*HistoryEle
}

// Every method create and insert DB.

// The number of binding card.
// Using for check, in create section.
const BINDINGCARD_LIMIT = 3

// Default validthru period (month)
const ACCOUNT_VALIDTHRU_PERIOD = 120

// Default display histoy lines
const DISPLAY_HISTORY_LINE = 30

// Prevent from overflow
func SafeAdd(a, b Money) (Money, bool) {
	if c := a + b; c >= a {
		return c, true
	}
	return 0, false
}

// You can add here Card valid checker.
// card pin/number form, limit etc...
func (a *Account) NewCard(pin CardPIN) (*Card, error) {
	// Check limit of the number of bindcard.
	if len(CardFromAccountNumber[a.Number])+1 > BINDINGCARD_LIMIT {
		return nil, fmt.Errorf("excess limit of card quantity (limit %d)", BINDINGCARD_LIMIT)
	}
	cardNumber := CardNumber(GenerateUUID())
	// Card is unique.
	if _, ok := SyncCardFromNumber[cardNumber]; ok {
		return nil, errors.New("card number is already exist")
	}
	t := time.Now()
	card := &Card{
		Number:    cardNumber,
		PIN:       pin,
		ValidThru: time.Date(t.Year(), t.Month()+CARD_VALIDTHRU_PERIOD, t.Day(), 0, 0, 0, 0, time.UTC),
	}
	// Insert DB
	InsertCardDB(a, card)
	return card, nil
}

func (a *Account) Subtract(m Money) error {
	if a.Balance < m {
		return errors.New("not enough balance")
	}
	a.Balance -= m
	return nil
}

func (a *Account) Add(m Money) error {
	v, ok := SafeAdd(a.Balance, m)
	if !ok {
		return errors.New("account will overflow")
	}
	a.Balance = v
	return nil
}

func (a *Account) Deposit(m Money) error {
	AccountMutex := SyncAccountFromNumber[a.Number].Mutex
	AccountMutex.Lock()
	err := a.Add(m)
	if err == nil {
		a.History = append(a.History, &HistoryEle{
			T:          time.Now(),
			Action:     "Deposit",
			Amount:     m,
			AccBalance: a.Balance,
		})
	}
	AccountMutex.Unlock()
	return err
}

func (a *Account) Withdrawal(m Money) error {
	AccountMutex := SyncAccountFromNumber[a.Number].Mutex
	AccountMutex.Lock()
	err := a.Subtract(m)
	if err == nil {
		a.History = append(a.History, &HistoryEle{
			T:          time.Now(),
			Action:     "Withdrawal",
			Amount:     m,
			AccBalance: a.Balance,
		})
	}
	AccountMutex.Unlock()
	return err
}

func (a *Account) Transfer(to AccountNumber, m Money) error {
	var err error
	fromAccount, toAccount := SyncAccountFromNumber[a.Number].Value, SyncAccountFromNumber[to].Value
	if fromAccount.Number == toAccount.Number {
		return errors.New("same account")
	}
	fromAccountMutex, toAccountMutex := SyncAccountFromNumber[a.Number].Mutex, SyncAccountFromNumber[to].Mutex
	fromClient, toClient := ClientFromAccountNumber[a.Number], ClientFromAccountNumber[to]
	fromAccountMutex.Lock()
	switch toAccountMutex.TryLock() {
	// Prevent from deadlock
	case false:
		fromAccountMutex.Unlock()
		return errors.New("someone is occupied")
	case true:

		if fromAccount.Balance < m {
			err = fmt.Errorf("%s %s not enough balance", fromClient.Name.Frist, fromClient.Name.Last)
			break
		}
		if _, ok := SafeAdd(toAccount.Balance, m); !ok {
			err = fmt.Errorf("%s %s account will overflow", toClient.Name.Frist, toClient.Name.Last)
			break
		}
		fromAccount.Balance -= m
		toAccount.Balance += m
		// Write history
		if err == nil {
			// FromAccount
			a.History = append(a.History, &HistoryEle{
				T:          time.Now(),
				Action:     "Transfer/Tx",
				Trader:     toAccount.Number,
				Amount:     m,
				AccBalance: a.Balance,
			})
			// ToAccount
			toAccount.History = append(toAccount.History, &HistoryEle{
				T:          time.Now(),
				Action:     "Transfer/Rx",
				Trader:     fromAccount.Number,
				Amount:     m,
				AccBalance: toAccount.Balance,
			})
		}
	}
	toAccountMutex.Unlock()
	fromAccountMutex.Unlock()
	if err != nil {
		return err
	}
	return nil
}

func (a *Account) GetHistory() []*HistoryEle {
	l := len(a.History)
	if l <= 30 {
		return a.History[:]
	}
	return a.History[l-DISPLAY_HISTORY_LINE:]
}

func (a *Account) PINCheck(pin string) error {
	if AccountPIN(pin) != a.PIN {
		return errors.New("wrong PIN")
	}
	return nil
}
