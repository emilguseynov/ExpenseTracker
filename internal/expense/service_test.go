package expense

import "testing"

func TestParseCents(t *testing.T) {
	tests := []struct {
		amount string
		want   int64
	}{
		{amount: "50.00", want: 5000},
		{amount: "50.10", want: 5010},
		{amount: "0.01", want: 1},
	}

	for _, tt := range tests {
		got, err := ParseCents(tt.amount)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != tt.want {
			t.Fatalf("wrong result: %v, want: %v", got, tt.want)
		}
	}
}

func TestParseCentsInvalid(t *testing.T) {
	tests := []string{"50.1", "50", "0", "0.00", "-1.00"}

	for _, tt := range tests {
		_, err := ParseCents(tt)
		if err == nil {
			t.Fatalf("got nil error, expected error: %v", ErrInvalidAmount)
		}
	}
}
