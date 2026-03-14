// Package money provides safe money handling for financial calculations
// Uses int64 to store amounts in smallest currency unit (cents)
// NEVER use float for money!
package money

import (
	"errors"
	"fmt"
)

var ErrCurrencyMismatch = errors.New("currency mismatch")

// Money represents a monetary amount with its currency
type Money struct {
	Amount   int64  // In smallest currency unit (e.g., cents for USD)
	Currency string // ISO 4217 currency code
}

// New creates a new Money instance
func New(amount int64, currency string) Money {
	return Money{Amount: amount, Currency: currency}
}

// Add returns the sum of two Money values
// Returns error if currencies don't match
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Sub returns the difference of two Money values
// Returns error if currencies don't match
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

// IsZero returns true if the amount is zero
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsNegative returns true if the amount is negative
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// IsPositive returns true if the amount is positive
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// Negate returns the negation of the money value
func (m Money) Negate() Money {
	return Money{
		Amount:   -m.Amount,
		Currency: m.Currency,
	}
}

// Equal returns true if two Money values are equal
func (m Money) Equal(other Money) bool {
	return m.Currency == other.Currency && m.Amount == other.Amount
}

// LessThan returns true if m < other
func (m Money) LessThan(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}
	return m.Amount < other.Amount
}

// GreaterThan returns true if m > other
func (m Money) GreaterThan(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}
	return m.Amount > other.Amount
}

// String returns a string representation of the money value
func (m Money) String() string {
	switch m.Currency {
	case "USD", "EUR", "GBP":
		return fmt.Sprintf("%s %.2f", m.Currency, float64(m.Amount)/100)
	case "JPY":
		return fmt.Sprintf("%s %d", m.Currency, m.Amount)
	default:
		return fmt.Sprintf("%s %d (minor units)", m.Currency, m.Amount)
	}
}

// ToMajorUnits returns the amount in major currency units as float64
// WARNING: Only use for display purposes, NEVER for calculations
func (m Money) ToMajorUnits() float64 {
	return float64(m.Amount) / 100
}

// Zero returns a zero-value Money for the given currency
func Zero(currency string) Money {
	return Money{Amount: 0, Currency: currency}
}
