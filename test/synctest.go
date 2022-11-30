package test

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/yusong-offx/myATM/atm"
	"github.com/yusong-offx/myATM/bank"
)

type Sub_1 struct {
	FromAccount *bank.Account
	ToAccount   *bank.Account
	FromCard    *bank.Card
	ToCard      *bank.Card
}

const (
	TEST_VOL = 10
	ATM_VOL  = 3
)

func SyncTest() {
	var err error
	fmt.Println("Sync Test Start")
	InitTest()
	if err = sub_1(); err != nil {
		logger.Fatal(err)
	}
	if err = sub_2(); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("Finish Test Check logfile..")
}

func InitTest() {
	fmt.Println("Init test...")
	// Generate
	// ATM
	for i := 0; i < ATM_VOL; i++ {
		atmlist = append(atmlist, atm.NewATM())
	}
	// Client <-- Account <-- Card
	for i := 0; i < TEST_VOL; i++ {
		client, _ := bank.NewClient(&bank.Client{
			SSN: bank.SSN(bank.GenerateUUID()),
		})
		clientlist = append(clientlist, client)
		acc, _ := client.NewAccount(PWD)
		accountlist = append(accountlist, acc)
		card, _ := acc.NewCard(PWD)
		cardlist = append(cardlist, card)
		// This will be written
		card.Deposit(1000)
	}
	// skip close
	logfile, _ := os.Create("./test/synctest.log")
	logger = log.New(logfile, "", log.Ldate)
}

// Check action order.
func checkHistory() error {
	for _, wrap_account := range bank.SyncAccountFromNumber {
		history := wrap_account.Value.History
		for i := 1; i < len(history); i++ {
			if history[i].T.Sub(history[i-1].T).Nanoseconds() <= 0 {
				return errors.New("history error")
			}
		}
	}
	return nil
}

// Check total money
// ATM total money 		= ATM_BASIC_DEPOSIT(10000) * ATM_VOL(2)
// Client total money	= TEST_VOL * 1000(first deposit)
func checkMoney() error {
	// ATM total money
	var totalATMMoney uint64 = 0
	for _, a := range atmlist {
		totalATMMoney += uint64(a.Cash)
	}
	if atm.ATM_BASIC_DEPOSIT*ATM_VOL != totalATMMoney {
		return fmt.Errorf("total atm money error (%d)", totalATMMoney)
	}
	// Client total money
	var totalClientMoney uint64 = 0
	for _, a := range accountlist {
		totalClientMoney += uint64(a.Balance)
	}
	if TEST_VOL*1000 != totalClientMoney {
		return fmt.Errorf("total atm money error (%d)", totalClientMoney)
	}
	return nil
}

// One ATM
// Account and its bind Card access same time.
func sub_1() error {
	fmt.Println("Test 1...")
	logger.Println("Test 1 Start")
	oneATM := atmlist[0]
	wg := sync.WaitGroup{}
	datas := sub_1_data()
	for _, data := range datas {
		// No transfer occupied error
		// Cause use oneATM
		for i := 0; i < 100; i++ {
			wg.Add(2)
			// Account
			go func(from, to *bank.Account) {
				defer wg.Done()
				if err := oneATM.Read(from, PWD); err != nil {
					logger.Println(err)
					return
				}
				if err := oneATM.Transfer(to.Number, 10); err != nil {
					logger.Println(err)
					return
				}
				logger.Println(from.Number, from.History[len(from.History)-1])
				oneATM.Return()
			}(data.FromAccount, data.ToAccount)
			// Card
			go func(from, to *bank.Card) {
				defer wg.Done()
				if err := oneATM.Read(from, PWD); err != nil {
					logger.Println(err)
					return
				}
				if err := oneATM.Transfer(bank.AccountFromCardNumber[to.Number].Number, 10); err != nil {
					logger.Println(err)
					return
				}
				history := oneATM.GetHistory()
				logger.Println(from.Number, history[len(history)-1])
				oneATM.Return()
			}(data.FromCard, data.ToCard)
		}
	}
	wg.Wait()
	if err := checkHistory(); err != nil {
		return err
	}
	// If you get history through ATM maximum shown 30.
	// Under 30 then all
	oneATM.Read(accountlist[0], PWD)
	if len(oneATM.GetHistory()) > 30 {
		return errors.New("ATM GetHistory error")
	}
	oneATM.Return()
	// Money check
	if !(datas[0].FromAccount.Balance == 1000 &&
		datas[0].ToAccount.Balance == 1000) {
		return errors.New("total money error")
	}
	return nil
}

func sub_1_data() []Sub_1 {
	testlist := make([]Sub_1, 2)
	testlist[0] = Sub_1{
		FromAccount: accountlist[0],
		ToAccount:   accountlist[1],
		FromCard:    cardlist[0],
		ToCard:      cardlist[1],
	}
	testlist[1] = Sub_1{
		FromAccount: accountlist[1],
		ToAccount:   accountlist[0],
		FromCard:    cardlist[1],
		ToCard:      cardlist[0],
	}
	return testlist
}

// Using random ATM 2(ATM_VOL)
// transfer, withdrawal, deposit 5(TEST_VOL) times
// each all account.
func sub_2() error {
	var err error
	fmt.Println("Test 2...")
	logger.Println("Test 2 Start")
	wg := sync.WaitGroup{}
	rand.Seed(time.Now().UnixNano())
	for same, account := range accountlist {
		wg.Add(1)
		go func(a *bank.Account, number int) {
			defer wg.Done()
			// Transfer
			for i := 0; i < 5; i++ {
				nowATM := atmlist[rand.Intn(ATM_VOL)]
				toAccount := accountlist[getToAccount(number)]
				nowATM.Read(a, PWD)
				// Must Transfer
				for {
					if err := nowATM.Transfer(toAccount.Number, 10); err == nil {
						break
					}
					// Reduce busy wait
					time.Sleep(time.Microsecond * 100)
				}
				logger.Println(a.Number, a.History[len(a.History)-1])
				nowATM.Return()
			}
			// Withdrawal
			for i := 0; i < 5; i++ {
				nowATM := atmlist[rand.Intn(ATM_VOL)]
				nowATM.Read(a, PWD)
				if err := nowATM.Withdrawal(1); err != nil {
					logger.Println(err)
				}
				logger.Println(a.Number, a.History[len(a.History)-1])
				nowATM.Return()
			}
			// Depoist
			for i := 0; i < 5; i++ {
				nowATM := atmlist[rand.Intn(ATM_VOL)]
				nowATM.Read(a, PWD)
				if err := nowATM.Deposit(1); err != nil {
					logger.Println(err)
				}
				logger.Println(a.Number, a.History[len(a.History)-1])
				nowATM.Return()
			}
		}(account, same)
	}
	wg.Wait()
	// Write ATM money
	logger.Println("ATM Cash Storage")
	for i, a := range atmlist {
		logger.Println(i, a.ID, a.Cash)
	}
	if err = checkHistory(); err != nil {
		return err
	}
	if err = checkMoney(); err != nil {
		return err
	}
	return nil
}

func getToAccount(same int) int {
	for {
		ret := rand.Intn(TEST_VOL)
		if ret != same {
			return ret
		}
	}
}
