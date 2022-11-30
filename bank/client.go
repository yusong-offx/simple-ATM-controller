package bank

import (
	"errors"
	"time"
)

type SSN string

type Name struct {
	Frist string
	Last  string
}

type Address struct {
	Street  string
	City    string
	ZipCode string
}

type Client struct {
	SSN         SSN
	Name        Name
	Gender      bool
	Birth       time.Time
	Country     string
	PhoneNumber string
	Email       string
	Address     Address
}

// Every method create and insert DB.

// You can add here Client valid checker.
// e.g. SSN etc....
func NewClient(client *Client) (*Client, error) {
	// Client is unique.
	if _, ok := SyncClientFromSSN[client.SSN]; ok {
		return nil, errors.New("client is already exist")
	}

	// Insert DB
	InsertClientDB(client)
	return client, nil
}

// You can add here account valid checker.
// e.g. pin/number form, limit(quantity) etc...
func (c *Client) NewAccount(pin AccountPIN) (*Account, error) {
	acccountNumber := AccountNumber(GenerateUUID())
	// Account is unique.
	if _, ok := SyncAccountFromNumber[acccountNumber]; ok {
		return nil, errors.New("account number is already exist")
	}
	t := time.Now()
	account := &Account{
		Number:    acccountNumber,
		PIN:       pin,
		ValidThru: time.Date(t.Year(), t.Month()+ACCOUNT_VALIDTHRU_PERIOD, t.Day(), 0, 0, 0, 0, time.UTC),
	}
	// Insert DB
	InsertAccountDB(c, account)
	return account, nil
}
