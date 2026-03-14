package ledger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAccountNotFound = errors.New("account not found")
var ErrJournalNotFound = errors.New("journal not found")
var ErrDoubleEntryViolation = errors.New("double-entry violation: debits must equal credits")

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateAccount(ctx context.Context, account *Account) error {
	query := `
		INSERT INTO accounts (id, code, name, type, currency, normal_balance, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		account.ID,
		account.Code,
		account.Name,
		account.Type,
		account.Currency,
		account.NormalBalance,
		account.ParentID,
		account.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	query := `SELECT id, code, name, type, currency, normal_balance, parent_id, created_at FROM accounts WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var account Account
	err := row.Scan(
		&account.ID,
		&account.Code,
		&account.Name,
		&account.Type,
		&account.Currency,
		&account.NormalBalance,
		&account.ParentID,
		&account.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAccountNotFound
	}
	return &account, err
}

func (r *PostgresRepository) GetAccountByCode(ctx context.Context, code string) (*Account, error) {
	query := `SELECT id, code, name, type, currency, normal_balance, parent_id, created_at FROM accounts WHERE code = $1`
	row := r.pool.QueryRow(ctx, query, code)

	var account Account
	err := row.Scan(
		&account.ID,
		&account.Code,
		&account.Name,
		&account.Type,
		&account.Currency,
		&account.NormalBalance,
		&account.ParentID,
		&account.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAccountNotFound
	}
	return &account, err
}

func (r *PostgresRepository) ListAccounts(ctx context.Context) ([]Account, error) {
	query := `SELECT id, code, name, type, currency, normal_balance, parent_id, created_at FROM accounts ORDER BY code`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var account Account
		err := rows.Scan(
			&account.ID,
			&account.Code,
			&account.Name,
			&account.Type,
			&account.Currency,
			&account.NormalBalance,
			&account.ParentID,
			&account.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (r *PostgresRepository) CreateJournal(ctx context.Context, journal *Journal) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert journal header
	journalQuery := `
		INSERT INTO journals (id, reference_type, reference_id, description, posted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(ctx, journalQuery,
		journal.ID,
		journal.ReferenceType,
		journal.ReferenceID,
		journal.Description,
		journal.PostedAt,
		journal.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create journal: %w", err)
	}

	// Insert journal lines
	lineQuery := `
		INSERT INTO journal_lines (id, journal_id, account_id, amount, direction, currency, memo, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	for _, line := range journal.Lines {
		_, err = tx.Exec(ctx, lineQuery,
			line.ID,
			line.JournalID,
			line.AccountID,
			line.Amount,
			line.Direction,
			line.Currency,
			line.Memo,
			line.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create journal line: %w", err)
		}
	}

	// The DB trigger will enforce double-entry
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit journal: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetJournalByID(ctx context.Context, id uuid.UUID) (*Journal, error) {
	// Get journal header
	query := `SELECT id, reference_type, reference_id, description, posted_at, created_at FROM journals WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var journal Journal
	err := row.Scan(
		&journal.ID,
		&journal.ReferenceType,
		&journal.ReferenceID,
		&journal.Description,
		&journal.PostedAt,
		&journal.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrJournalNotFound
	}
	if err != nil {
		return nil, err
	}

	// Get journal lines
	linesQuery := `
		SELECT jl.id, jl.journal_id, jl.account_id, jl.amount, jl.direction, jl.currency, jl.memo, jl.created_at,
		       a.id, a.code, a.name, a.type, a.currency, a.normal_balance
		FROM journal_lines jl
		JOIN accounts a ON jl.account_id = a.id
		WHERE jl.journal_id = $1
		ORDER BY jl.created_at
	`
	lines, err := r.pool.Query(ctx, linesQuery, id)
	if err != nil {
		return nil, err
	}
	defer lines.Close()

	for lines.Next() {
		var line JournalLine
		var acc Account
		err := lines.Scan(
			&line.ID,
			&line.JournalID,
			&line.AccountID,
			&line.Amount,
			&line.Direction,
			&line.Currency,
			&line.Memo,
			&line.CreatedAt,
			&acc.ID,
			&acc.Code,
			&acc.Name,
			&acc.Type,
			&acc.Currency,
			&acc.NormalBalance,
		)
		if err != nil {
			return nil, err
		}
		line.Account = &acc
		journal.Lines = append(journal.Lines, line)
	}

	return &journal, nil
}

func (r *PostgresRepository) ListJournals(ctx context.Context, limit, offset int) ([]Journal, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `SELECT id, reference_type, reference_id, description, posted_at, created_at 
		FROM journals ORDER BY posted_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var journals []Journal
	for rows.Next() {
		var j Journal
		err := rows.Scan(&j.ID, &j.ReferenceType, &j.ReferenceID, &j.Description, &j.PostedAt, &j.CreatedAt)
		if err != nil {
			return nil, err
		}
		journals = append(journals, j)
	}
	return journals, nil
}

func (r *PostgresRepository) GetJournalsByReference(ctx context.Context, refType string, refID uuid.UUID) ([]Journal, error) {
	query := `SELECT id, reference_type, reference_id, description, posted_at, created_at 
		FROM journals WHERE reference_type = $1 AND reference_id = $2 ORDER BY posted_at DESC`
	rows, err := r.pool.Query(ctx, query, refType, refID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var journals []Journal
	for rows.Next() {
		var j Journal
		err := rows.Scan(&j.ID, &j.ReferenceType, &j.ReferenceID, &j.Description, &j.PostedAt, &j.CreatedAt)
		if err != nil {
			return nil, err
		}
		journals = append(journals, j)
	}
	return journals, nil
}

func (r *PostgresRepository) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (*AccountBalance, error) {
	account, err := r.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN direction = 'DEBIT' THEN amount ELSE 0 END), 0) as debit_total,
			COALESCE(SUM(CASE WHEN direction = 'CREDIT' THEN amount ELSE 0 END), 0) as credit_total
		FROM journal_lines
		WHERE account_id = $1
	`
	row := r.pool.QueryRow(ctx, query, accountID)

	var debitTotal, creditTotal int64
	err = row.Scan(&debitTotal, &creditTotal)
	if err != nil {
		return nil, err
	}

	// Calculate balance based on account type
	var balance int64
	if account.NormalBalance == BalanceTypeDebit {
		balance = debitTotal - creditTotal
	} else {
		balance = creditTotal - debitTotal
	}

	return &AccountBalance{
		AccountID:   accountID,
		AccountCode: account.Code,
		AccountName: account.Name,
		DebitTotal:  debitTotal,
		CreditTotal: creditTotal,
		Balance:     balance,
		Currency:    account.Currency,
	}, nil
}

func (r *PostgresRepository) GetAllAccountBalances(ctx context.Context) ([]AccountBalance, error) {
	query := `
		SELECT 
			a.id, a.code, a.name,
			COALESCE(SUM(CASE WHEN jl.direction = 'DEBIT' THEN jl.amount ELSE 0 END), 0) as debit_total,
			COALESCE(SUM(CASE WHEN jl.direction = 'CREDIT' THEN jl.amount ELSE 0 END), 0) as credit_total,
			a.currency
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		GROUP BY a.id, a.code, a.name, a.currency, a.normal_balance
		ORDER BY a.code
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []AccountBalance
	for rows.Next() {
		var b AccountBalance
		var accountName string
		err := rows.Scan(&b.AccountID, &b.AccountCode, &accountName, &b.DebitTotal, &b.CreditTotal, &b.Currency)
		if err != nil {
			return nil, err
		}
		b.AccountName = accountName

		// Get account to determine balance type
		account, err := r.GetAccountByID(ctx, b.AccountID)
		if err != nil {
			return nil, err
		}

		if account.NormalBalance == BalanceTypeDebit {
			b.Balance = b.DebitTotal - b.CreditTotal
		} else {
			b.Balance = b.CreditTotal - b.DebitTotal
		}

		balances = append(balances, b)
	}
	return balances, nil
}
