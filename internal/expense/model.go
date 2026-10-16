package expense

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

var ErrInvalidAmount = errors.New("invalid amount")

func ParseCents(amount string) (int64, error) {
	whole, frac, ok := strings.Cut(amount, ".")
	if !ok || len(frac) != 2 || strings.Contains(frac, ".") {
		return 0, ErrInvalidAmount
	}
	if !digitsOnly(whole) || !digitsOnly(frac) {
		return 0, ErrInvalidAmount
	}

	dollars, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}
	cents, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmount
	}

	total := dollars*100 + cents
	if total <= 0 {
		return 0, ErrInvalidAmount
	}
	return total, nil
}

func digitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
