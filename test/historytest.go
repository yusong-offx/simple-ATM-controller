package test

import (
	"log"
	"os"

	"github.com/yusong-offx/myATM/atm"
	"github.com/yusong-offx/myATM/bank"
)

const PWD = "0000"

var (
	clientlist  []*bank.Client
	accountlist []*bank.Account
	cardlist    []*bank.Card
	atmlist     = []*atm.ATM{}
	logger      *log.Logger
)

// (client-account-card) 2 sets
func HistoryTest() {
	logfile, _ := os.Create("./test/historytest.log")
	logger = log.New(logfile, "", log.Ldate)

	for i := 0; i < 2; i++ {
		client, _ := bank.NewClient(&bank.Client{
			SSN: bank.SSN(bank.GenerateUUID()),
		})
		clientlist = append(clientlist, client)
		acc, _ := client.NewAccount(PWD)
		accountlist = append(accountlist, acc)
		card, _ := acc.NewCard(PWD)
		cardlist = append(cardlist, card)
		// This will be written
		card.Deposit(1500)
		card.Withdrawal(500)
	}
	// This will be written
	for i := 0; i < 5; i++ {
		cardlist[0].Transfer(accountlist[1].Number, 10)
	}

	//// Display ////
	// Func GetHistory test
	for k, v := range bank.SyncAccountFromNumber {
		logger.Println("Account Number : ", k)
		for _, l := range v.Value.History {
			logger.Println(l)
		}
		logger.Println()
	}
	// Using ATM
	myatm := atm.NewATM()
	myatm.Read(cardlist[0], PWD)
	logger.Println("Using ATM Card Number : ", cardlist[0].Number, "bind account", bank.AccountFromCardNumber[cardlist[0].Number].Number)
	for _, l := range myatm.GetHistory() {
		logger.Println(l)
	}
	myatm.Return()
}
