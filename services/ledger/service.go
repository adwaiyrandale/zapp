package ledger

import (
	"context"
	"errors"
	"time"

	"github.com/adwaiy/zap/pkg/money"
	"github.com/google/uuid"
)

var (
	ErrInvalidAmount    = errors.New("amount must be positive")
	ErrCurrencyMismatch = errors.New("currency mismatch between entries")
	ErrAccountRequired  = errors.New("at least one debit and one credit entry required")
	ErrMissingAccount   = errors.New("account not found")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type PostJournalInput struct {
	ReferenceType string
	ReferenceID   uuid.UUID
	Description   string
	Entries       []PostEntry
}

type PostEntry struct {
	AccountCode string
	Amount      int64  // in cents
	Direction   string // DEBIT or CREDIT
	Currency    string
	Memo        string
}

func (s *Service) PostJournal(ctx context.Context, input *PostJournalInput) (*Journal, error) {
	// Validate input
	if len(input.Entries) < 2 {
		return nil, ErrAccountRequired
	}

	// Validate entries and get account IDs
	accountIDs := make(map[string]uuid.UUID)
	var entries []JournalLine
	var firstCurrency string

	hasDebit := false
	hasCredit := false

	for _, entry := range input.Entries {
		// Validate amount
		if entry.Amount <= 0 {
			return nil, ErrInvalidAmount
		}

		// Get account
		account, err := s.repo.GetAccountByCode(ctx, entry.AccountCode)
		if err != nil {
			return nil, errors.Join(ErrMissingAccount, err)
		}

		// Validate direction
		if entry.Direction != "DEBIT" && entry.Direction != "CREDIT" {
			return nil, errors.New("direction must be DEBIT or CREDIT")
		}

		if entry.Direction == "DEBIT" {
			hasDebit = true
		} else {
			hasCredit = true
		}

		// Validate currency
		if firstCurrency == "" {
			firstCurrency = entry.Currency
		}
		if entry.Currency != firstCurrency {
			return nil, ErrCurrencyMismatch
		}

		accountIDs[entry.AccountCode] = account.ID
		entries = append(entries, JournalLine{
			ID:        uuid.New(),
			AccountID: account.ID,
			Amount:    entry.Amount,
			Direction: entry.Direction,
			Currency:  entry.Currency,
			Memo:      entry.Memo,
			CreatedAt: time.Now().UTC(),
		})
	}

	if !hasDebit || !hasCredit {
		return nil, ErrAccountRequired
	}

	// Calculate totals for validation
	var totalDebits, totalCredits int64
	for _, entry := range entries {
		if entry.Direction == "DEBIT" {
			totalDebits += entry.Amount
		} else {
			totalCredits += entry.Amount
		}
	}

	// Enforce double-entry at application level too
	if totalDebits != totalCredits {
		return nil, errors.Join(ErrDoubleEntryViolation,
			errors.New("debits and credits must be equal"))
	}

	// Create journal
	journal := &Journal{
		ID:            uuid.New(),
		ReferenceType: input.ReferenceType,
		ReferenceID:   input.ReferenceID,
		Description:   input.Description,
		PostedAt:      time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
		Lines:         entries,
	}

	// Set journal ID on all lines
	for i := range journal.Lines {
		journal.Lines[i].JournalID = journal.ID
	}

	// Save to database (DB trigger will also enforce double-entry)
	if err := s.repo.CreateJournal(ctx, journal); err != nil {
		return nil, err
	}

	return journal, nil
}

func (s *Service) GetAccount(ctx context.Context, code string) (*Account, error) {
	return s.repo.GetAccountByCode(ctx, code)
}

func (s *Service) ListAccounts(ctx context.Context) ([]Account, error) {
	return s.repo.ListAccounts(ctx)
}

func (s *Service) GetAccountBalance(ctx context.Context, code string) (*AccountBalance, error) {
	account, err := s.repo.GetAccountByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.repo.GetAccountBalance(ctx, account.ID)
}

func (s *Service) GetAllBalances(ctx context.Context) ([]AccountBalance, error) {
	return s.repo.GetAllAccountBalances(ctx)
}

func (s *Service) GetJournal(ctx context.Context, id uuid.UUID) (*Journal, error) {
	return s.repo.GetJournalByID(ctx, id)
}

func (s *Service) ListJournals(ctx context.Context, limit, offset int) ([]Journal, error) {
	return s.repo.ListJournals(ctx, limit, offset)
}

// Convenience methods for common transactions

func (s *Service) RecordPayment(ctx context.Context, paymentID uuid.UUID, amount money.Money, description string) (*Journal, error) {
	return s.PostJournal(ctx, &PostJournalInput{
		ReferenceType: "PAYMENT",
		ReferenceID:   paymentID,
		Description:   description,
		Entries: []PostEntry{
			{
				AccountCode: "1000", // Cash
				Amount:      amount.Amount,
				Direction:   "DEBIT",
				Currency:    amount.Currency,
			},
			{
				AccountCode: "2000", // Revenue
				Amount:      amount.Amount,
				Direction:   "CREDIT",
				Currency:    amount.Currency,
			},
		},
	})
}

func (s *Service) RecordRefund(ctx context.Context, refundID uuid.UUID, amount money.Money, description string) (*Journal, error) {
	return s.PostJournal(ctx, &PostJournalInput{
		ReferenceType: "REFUND",
		ReferenceID:   refundID,
		Description:   description,
		Entries: []PostEntry{
			{
				AccountCode: "2100", // Refunds Payable
				Amount:      amount.Amount,
				Direction:   "DEBIT",
				Currency:    amount.Currency,
			},
			{
				AccountCode: "1000", // Cash
				Amount:      amount.Amount,
				Direction:   "CREDIT",
				Currency:    amount.Currency,
			},
		},
	})
}

func (s *Service) RecordFee(ctx context.Context, feeID uuid.UUID, amount money.Money, description string) (*Journal, error) {
	return s.PostJournal(ctx, &PostJournalInput{
		ReferenceType: "FEE",
		ReferenceID:   feeID,
		Description:   description,
		Entries: []PostEntry{
			{
				AccountCode: "3000", // Payment Processing Fees
				Amount:      amount.Amount,
				Direction:   "DEBIT",
				Currency:    amount.Currency,
			},
			{
				AccountCode: "1000", // Cash
				Amount:      amount.Amount,
				Direction:   "CREDIT",
				Currency:    amount.Currency,
			},
		},
	})
}
