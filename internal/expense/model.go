package expense

import (
	"fmt"
	"time"
)

type Expense struct {
	ID          int64
	AmountCents int64
	Category    string
	Description string
	CreatedAt   time.Time
}

func FormatCents(cents int64) string {
	negative := cents < 0
	dollars := cents / 100
	centsPart := cents % 100

	if negative {
		dollars = -dollars
		centsPart = -centsPart
	}

	res := fmt.Sprintf("%d.%02d", dollars, centsPart)

	if negative {
		res = "-" + res
	}

	return res
}
