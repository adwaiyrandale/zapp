package ledger

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET"
	AccountTypeLiability AccountType = "LIABILITY"
	AccountTypeEquity    AccountType = "EQUITY"
	AccountTypeRevenue   AccountType = "REVENUE"
	AccountTypeExpense   AccountType = "EXPENSE"
)

type BalanceType string

const (
	BalanceTypeDebit  BalanceType = "DEBIT"
	BalanceTypeCredit BalanceType = "CREDIT"
)

type Account struct {
	ID            uuid.UUID   `json:"id"`
	Code          string      `json:"code"`
	Name          string      `json:"name"`
	Type          AccountType `json:"type"`
	Currency      string      `json:"currency"`
	NormalBalance BalanceType `json:"normal_balance"`
	ParentID      *uuid.UUID  `json:"parent_id,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

type Journal struct {
	ID            uuid.UUID     `json:"id"`
	ReferenceType string        `json:"reference_type"`
	ReferenceID   uuid.UUID     `json:"reference_id"`
	Description   string        `json:"description,omitempty"`
	PostedAt      time.Time     `json:"posted_at"`
	CreatedAt     time.Time     `json:"created_at"`
	Lines         []JournalLine `json:"lines,omitempty"`
}

type JournalLine struct {
	ID        uuid.UUID `json:"id"`
	JournalID uuid.UUID `json:"journal_id"`
	AccountID uuid.UUID `json:"account_id"`
	Amount    int64     `json:"amount"`
	Direction string    `json:"direction"` // DEBIT or CREDIT
	Currency  string    `json:"currency"`
	Memo      string    `json:"memo,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Joined data
	Account *Account `json:"account,omitempty"`
}

type AccountBalance struct {
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code"`
	AccountName string    `json:"account_name"`
	DebitTotal  int64     `json:"debit_total"`
	CreditTotal int64     `json:"credit_total"`
	Balance     int64     `json:"balance"`
	Currency    string    `json:"currency"`
}

type Repository interface {
	// Account operations
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetAccountByCode(ctx context.Context, code string) (*Account, error)
	ListAccounts(ctx context.Context) ([]Account, error)

	// Journal operations
	CreateJournal(ctx context.Context, journal *Journal) error
	GetJournalByID(ctx context.Context, id uuid.UUID) (*Journal, error)
	ListJournals(ctx context.Context, limit, offset int) ([]Journal, error)
	GetJournalsByReference(ctx context.Context, refType string, refID uuid.UUID) ([]Journal, error)

	// Balance operations
	GetAccountBalance(ctx context.Context, accountID uuid.UUID) (*AccountBalance, error)
	GetAllAccountBalances(ctx context.Context) ([]AccountBalance, error)
}
