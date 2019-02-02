package stdout

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type Datum struct {
	Name   string
	Amount decimal.Decimal
}

type Table struct {
	Caption string
	Headers []string
	Data    []Datum
}

func NewTable(caption string, nameHeader, amountHeader string) *Table {
	return &Table{
		Caption: caption,
		Headers: []string{nameHeader, amountHeader},
		Data:    []Datum{},
	}
}

func (t *Table) AddRow(name string, amount decimal.Decimal) {
	t.Data = append(t.Data, Datum{name, amount})
}

func (t *Table) Print() {
	maxNameLength := 0
	for _, d := range t.Data {
		if len(d.Name) > maxNameLength {
			maxNameLength = len(d.Name)
		}
	}

	maxAmountLength := 10
	maxLength := maxNameLength + 1 + maxAmountLength
	rowFormat := "%-" + strconv.Itoa(maxNameLength) + "s %" + strconv.Itoa(maxAmountLength) + "s\n"

	fmt.Println("\n")
	fmt.Printf(t.Caption)
	captionLength := len(t.Caption)
	for captionLength < maxLength {
		fmt.Printf("=")
		captionLength++
	}
	fmt.Println("")

	// headers
	fmt.Printf(rowFormat, t.Headers[0], t.Headers[1])

	// header separator bar
	i := 0
	for i < maxLength {
		fmt.Printf("-")
		i++
	}
	fmt.Println("")

	// data rows
	for _, d := range t.Data {
		fmt.Printf(rowFormat, d.Name, prettyAmount(d.Amount))
	}

	// ending bar
	i = 0
	for i < maxLength {
		fmt.Printf("=")
		i++
	}
	fmt.Println("\n")
}

func prettyAmount(amount decimal.Decimal) string {
	amountStr := amount.StringFixed(2)
	if !strings.HasSuffix(amountStr, "00") {
		return amountStr
	}
	return strings.TrimSuffix(amountStr, "00") + "  "
}
