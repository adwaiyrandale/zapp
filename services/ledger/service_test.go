package ledger

import (
	"context"
	"testing"

	"github.com/adwaiy/zap/pkg/money"
	"github.com/google/uuid"
)

type mockRepository struct {
	accounts map[string]*Account
	journals map[uuid.UUID]*Journal
	balances map[uuid.UUID]*AccountBalance
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		accounts: map[string]*Account{
			"1000": {ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Code: "1000", Name: "Cash", Type: AccountTypeAsset, Currency: "USD", NormalBalance: BalanceTypeDebit},
			"2000": {ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Code: "2000", Name: "Revenue", Type: AccountTypeRevenue, Currency: "USD", NormalBalance: BalanceTypeCredit},
			"2100": {ID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), Code: "2100", Name: "Refunds Payable", Type: AccountTypeLiability, Currency: "USD", NormalBalance: BalanceTypeCredit},
			"3000": {ID: uuid.MustParse("44444444-4444-4444-4444-444444444444"), Code: "3000", Name: "Payment Processing Fees", Type: AccountTypeExpense, Currency: "USD", NormalBalance: BalanceTypeDebit},
		},
		journals: make(map[uuid.UUID]*Journal),
		balances: make(map[uuid.UUID]*AccountBalance),
	}
}

func (m *mockRepository) CreateAccount(ctx context.Context, account *Account) error {
	m.accounts[account.Code] = account
	return nil
}

func (m *mockRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	for _, a := range m.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, ErrAccountNotFound
}

func (m *mockRepository) GetAccountByCode(ctx context.Context, code string) (*Account, error) {
	if a, ok := m.accounts[code]; ok {
		return a, nil
	}
	return nil, ErrAccountNotFound
}

func (m *mockRepository) ListAccounts(ctx context.Context) ([]Account, error) {
	var result []Account
	for _, a := range m.accounts {
		result = append(result, *a)
	}
	return result, nil
}

func (m *mockRepository) CreateJournal(ctx context.Context, journal *Journal) error {
	m.journals[journal.ID] = journal
	return nil
}

func (m *mockRepository) GetJournalByID(ctx context.Context, id uuid.UUID) (*Journal, error) {
	if j, ok := m.journals[id]; ok {
		return j, nil
	}
	return nil, ErrJournalNotFound
}

func (m *mockRepository) ListJournals(ctx context.Context, limit, offset int) ([]Journal, error) {
	var result []Journal
	for _, j := range m.journals {
		result = append(result, *j)
	}
	return result, nil
}

func (m *mockRepository) GetJournalsByReference(ctx context.Context, refType string, refID uuid.UUID) ([]Journal, error) {
	return nil, nil
}

func (m *mockRepository) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (*AccountBalance, error) {
	if b, ok := m.balances[accountID]; ok {
		return b, nil
	}
	return &AccountBalance{AccountID: accountID, Balance: 0}, nil
}

func (m *mockRepository) GetAllAccountBalances(ctx context.Context) ([]AccountBalance, error) {
	var result []AccountBalance
	for _, b := range m.balances {
		result = append(result, *b)
	}
	return result, nil
}

func TestPostJournal_DoubleEntry(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Test posting a balanced journal
	input := &PostJournalInput{
		ReferenceType: "TEST",
		ReferenceID:   uuid.New(),
		Description:   "Test payment",
		Entries: []PostEntry{
			{AccountCode: "1000", Amount: 1000, Direction: "DEBIT", Currency: "USD"},
			{AccountCode: "2000", Amount: 1000, Direction: "CREDIT", Currency: "USD"},
		},
	}

	journal, err := svc.PostJournal(context.Background(), input)
	if err != nil {
		t.Fatalf("PostJournal failed: %v", err)
	}

	if journal == nil {
		t.Fatal("Journal should not be nil")
	}

	if len(journal.Lines) != 2 {
		t.Errorf("Expected 2 journal lines, got %d", len(journal.Lines))
	}
}

func TestPostJournal_Unbalanced(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Test posting an unbalanced journal
	input := &PostJournalInput{
		ReferenceType: "TEST",
		ReferenceID:   uuid.New(),
		Description:   "Unbalanced test",
		Entries: []PostEntry{
			{AccountCode: "1000", Amount: 1000, Direction: "DEBIT", Currency: "USD"},
			{AccountCode: "2000", Amount: 500, Direction: "CREDIT", Currency: "USD"}, // Only 500!
		},
	}

	_, err := svc.PostJournal(context.Background(), input)
	if err == nil {
		t.Error("Expected error for unbalanced journal")
	}
}

func TestPostJournal_MissingAccount(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	input := &PostJournalInput{
		ReferenceType: "TEST",
		ReferenceID:   uuid.New(),
		Description:   "Missing account test",
		Entries: []PostEntry{
			{AccountCode: "1000", Amount: 1000, Direction: "DEBIT", Currency: "USD"},
			{AccountCode: "9999", Amount: 1000, Direction: "CREDIT", Currency: "USD"}, // Doesn't exist
		},
	}

	_, err := svc.PostJournal(context.Background(), input)
	if err == nil {
		t.Error("Expected error for missing account")
	}
}

func TestPostJournal_CurrencyMismatch(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	input := &PostJournalInput{
		ReferenceType: "TEST",
		ReferenceID:   uuid.New(),
		Description:   "Currency mismatch test",
		Entries: []PostEntry{
			{AccountCode: "1000", Amount: 1000, Direction: "DEBIT", Currency: "USD"},
			{AccountCode: "2000", Amount: 1000, Direction: "CREDIT", Currency: "EUR"}, // Different currency!
		},
	}

	_, err := svc.PostJournal(context.Background(), input)
	if err == nil {
		t.Error("Expected error for currency mismatch")
	}
}

func TestMoneyOperations(t *testing.T) {
	// Test using the money package
	amount1 := money.New(1000, "USD") // $10.00
	amount2 := money.New(500, "USD")  // $5.00

	sum, err := amount1.Add(amount2)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if sum.Amount != 1500 {
		t.Errorf("Expected 1500, got %d", sum.Amount)
	}

	diff, err := amount1.Sub(amount2)
	if err != nil {
		t.Fatalf("Sub failed: %v", err)
	}
	if diff.Amount != 500 {
		t.Errorf("Expected 500, got %d", diff.Amount)
	}

	// Test currency mismatch
	eur := money.New(1000, "EUR")
	_, err = amount1.Add(eur)
	if err != money.ErrCurrencyMismatch {
		t.Error("Expected currency mismatch error")
	}
}

func TestRecordPayment(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	paymentID := uuid.New()
	amount := money.New(10000, "USD") // $100.00

	journal, err := svc.RecordPayment(context.Background(), paymentID, amount, "Test payment")
	if err != nil {
		t.Fatalf("RecordPayment failed: %v", err)
	}

	if journal.ReferenceID != paymentID {
		t.Errorf("Expected reference ID %s, got %s", paymentID, journal.ReferenceID)
	}

	if journal.ReferenceType != "PAYMENT" {
		t.Errorf("Expected reference type PAYMENT, got %s", journal.ReferenceType)
	}
}
