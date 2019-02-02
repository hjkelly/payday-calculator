package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hjkelly/payday-calculator/stdout"
	"github.com/shopspring/decimal"
)

// MODELS =====================================================================

type Config struct {
	Paydays   []Payday  `json:"paydays"`
	Charities []Charity `json:"charities"`
	Expenses  []Expense `json:"expenses"`
	Bills     []Bill    `json:"bills"`
}

type Payday struct {
	Frequency string          `json:"frequency"`
	Amount    decimal.Decimal `json:"amount"`
}

const (
	SemimonthlyFirstHalf       = "semimonthlyFirstHalf"
	SemimonthlySecondHalf      = "semimonthlySecondHalf"
	SemimonthlyFirstHalfStart  = 1
	SemimonthlySecondHalfStart = 15
)

type Charity struct {
	Name       string          `json:"name"`
	Amount     decimal.Decimal `json:"amount"`
	Percentage decimal.Decimal `json:"percentage"`
}

type Expense struct {
	Name    string          `json:"name"`
	Amount  decimal.Decimal `json:"amount"`
	Balance decimal.Decimal `json:"balance"`
}

type Bill struct {
	Name     string          `json:"name"`
	Amount   decimal.Decimal `json:"amount"`
	DueOnDay int             `json:"dueOnDay"`
}

// MAIN PROCESS ===============================================================

func main() {
	fmt.Printf("Welcome to the payday calculator!\n")

	// determine config name
	configFile, err := filepath.Abs("config.json")
	if err != nil {
		fmt.Printf("Error while getting absolute path: %s\n", err.Error())
		os.Exit(1)
	}

	// load config
	config, err := openConfig(configFile)
	if err != nil {
		fmt.Printf("Couldn't load config from %s:\n%s\n", configFile, err.Error())
		os.Exit(1)
	}
	fmt.Printf("Successful config load: %s\n", configFile)
	// fmt.Printf("%+v\n", config)

	reader := bufio.NewReader(os.Stdin)

	// guess and confirm which half of the month we're doing the budget for
	halfMonth := guessHalfMonth(time.Now().Day())
	response := ""
	if halfMonth == SemimonthlyFirstHalf {
		fmt.Printf("It seems like you're preparing the budget for the FIRST HALF of the month. Is that right? Type 'y' or 'n'\n")
		response, _ = reader.ReadString('\n')
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "n") {
			halfMonth = SemimonthlySecondHalf
		}
	} else {
		fmt.Printf("It seems like you're preparing the budget for the SECOND HALF of the month. Is that right? Type 'y' or 'n'\n")
		response, _ = reader.ReadString('\n')
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "n") {
			halfMonth = SemimonthlyFirstHalf
		}
	}

	// FINALIZE PREVIOUS ==================================================

	// collect expensetotals for each category
	// TODO

	// load balances from previous file, if possible
	// TODO

	// fill up funds / steal from emergency fund for each category
	// TODO

	// save finalized balances to file
	// TODO 10

	// PLAN NEXT ==========================================================

	// confirm income
	income := decimal.NewFromFloat(0)
	for _, payday := range config.Paydays {
		income = income.Add(payday.Amount)
	}
	fmt.Printf("We expect your total income to put toward this budget is %s.\n", income)
	fmt.Printf("Provide a different amount if necessary. Press enter to accept.\n")
	rawIncome, _ := reader.ReadString('\n')
	rawIncome = strings.TrimSpace(rawIncome)
	if len(rawIncome) > 0 {
		parsedIncome, err := decimal.NewFromString(rawIncome)
		if err != nil {
			fmt.Printf("Couldn't understand your input. Sorry!\n%s\n", err.Error())
			os.Exit(2)
		}
		income = parsedIncome
	}

	// calculate giving/charities
	table := stdout.NewTable("CHARITIES", "Name", "Amount")
	charityTotal := decimal.NewFromFloat(0)
	for _, charity := range config.Charities {
		// calculate the amount
		actualAmount := charity.Amount
		if !charity.Percentage.IsZero() {
			actualAmount = income.Mul(charity.Percentage)
		}
		// increase the total
		charityTotal = charityTotal.Add(actualAmount)
		// get it ready for output
		table.AddRow(charity.Name, actualAmount)
	}
	table.Print()

	// get half of expenses ----------
	// loop through expenses, collecting the total amount and getting a table ready for output
	table = stdout.NewTable("EXPENSES", "Name", "Amount")
	expenseTotal := decimal.NewFromFloat(0)
	percentage := decimal.NewFromFloat(0.5)
	for _, expense := range config.Expenses {
		// half the amount since we're only budgeting for half the month
		budgetedItemAmount := expense.Amount.Mul(percentage)
		// increase the total
		expenseTotal = expenseTotal.Add(budgetedItemAmount)
		// get it ready for output
		table.AddRow(expense.Name, budgetedItemAmount)
	}
	table.Print()

	// confirm bill amounts
	// TODO 04

	// total bills for this half ---------
	// find which bills are relevant
	bills := make([]Bill, 0)
	for _, bill := range config.Bills {
		if halfMonth == SemimonthlyFirstHalf {
			if bill.DueOnDay >= SemimonthlyFirstHalfStart && bill.DueOnDay < SemimonthlySecondHalfStart {
				bills = append(bills, bill)
			}
		} else {
			if bill.DueOnDay < SemimonthlyFirstHalfStart || bill.DueOnDay >= SemimonthlySecondHalfStart {
				bills = append(bills, bill)
			}
		}
	}
	table = stdout.NewTable("BILLS DUE", "Name", "Amount")
	// loop through expenses, collecting the total amount and getting a table ready for output
	billTotal := decimal.NewFromFloat(0)
	for _, bill := range bills {
		billTotal = billTotal.Add(bill.Amount)
		table.AddRow(bill.Name, bill.Amount)
	}
	table.Print()

	// remind manual bills to pay
	// TODO 06

	// provide savings leftovers (or what needs to be stolen from emergency fund)
	balance := income.Sub(charityTotal).Sub(expenseTotal).Sub(billTotal)
	if balance.IsPositive() {
		fmt.Printf("This much is left over to save: %s\n", balance.String())
	} else if balance.IsNegative() {
		fmt.Printf("You don't have enough; can you steal this much from your emergency fund? %s\n", balance.Abs().String())
	} else {
		fmt.Printf("Wow, you broke even. Exactly. :|\n")
	}

	// save results to a file
	// TODO 05

	os.Exit(0)
}

// HELPER FUNCTIONS ===========================================================

func openConfig(filepath string) (Config, error) {
	config := Config{}

	// Open our jsonFile
	jsonFile, err := os.Open(filepath)
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	// if we os.Open returns an error then handle it
	if err != nil {
		return config, err
	}

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal([]byte(byteValue), &config)

	return config, err
}

func guessHalfMonth(day int) string {
	if day >= 10 && day < 20 {
		return SemimonthlySecondHalf
	}
	return SemimonthlyFirstHalf
}
