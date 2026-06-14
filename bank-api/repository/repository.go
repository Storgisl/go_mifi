package repository

import (
    "database/sql"
    "time"

    "bank-api/models"
    _ "github.com/lib/pq"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(connStr string) (*Repository, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
    return r.db.Close()
}

func (r *Repository) BeginTx() (*sql.Tx, error) {
    return r.db.Begin()
}

func (r *Repository) UpdateAccountBalanceTx(tx *sql.Tx, id int, newBalance float64) error {
    _, err := tx.Exec(`UPDATE accounts SET balance=$1 WHERE id=$2`, newBalance, id)
    return err
}

func (r *Repository) CreateTransactionTx(tx *sql.Tx, t *models.Transaction) (int, error) {
    var id int
    query := `INSERT INTO transactions (from_account_id, to_account_id, amount, type) VALUES ($1,$2,$3,$4) RETURNING id`
    err := tx.QueryRow(query, t.FromAccountID, t.ToAccountID, t.Amount, t.Type).Scan(&id)
    return id, err
}

func (r *Repository) CreateUser(user *models.User) (int, error) {
    var id int
    query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
    err := r.db.QueryRow(query, user.Username, user.Email, user.PasswordHash).Scan(&id)
    return id, err
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
    var u models.User
    query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email=$1`
    err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &u, err
}

func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
    var u models.User
    query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username=$1`
    err := r.db.QueryRow(query, username).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &u, err
}

func (r *Repository) GetUserByID(id int) (*models.User, error) {
    var u models.User
    query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id=$1`
    err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &u, err
}

func (r *Repository) CreateAccount(account *models.Account) (int, error) {
    var id int
    query := `INSERT INTO accounts (user_id, balance, currency) VALUES ($1, $2, $3) RETURNING id`
    err := r.db.QueryRow(query, account.UserID, account.Balance, account.Currency).Scan(&id)
    return id, err
}

func (r *Repository) GetAccountByID(id int) (*models.Account, error) {
    var acc models.Account
    query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id=$1`
    err := r.db.QueryRow(query, id).Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &acc, err
}

func (r *Repository) GetAccountsByUserID(userID int) ([]models.Account, error) {
    rows, err := r.db.Query(`SELECT id, user_id, balance, currency, created_at FROM accounts WHERE user_id=$1`, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var accounts []models.Account
    for rows.Next() {
        var a models.Account
        if err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.Currency, &a.CreatedAt); err != nil {
            return nil, err
        }
        accounts = append(accounts, a)
    }
    return accounts, nil
}

func (r *Repository) UpdateAccountBalance(id int, newBalance float64) error {
    _, err := r.db.Exec(`UPDATE accounts SET balance=$1 WHERE id=$2`, newBalance, id)
    return err
}

func (r *Repository) CreateTransaction(tx *models.Transaction) (int, error) {
    var id int
    query := `INSERT INTO transactions (from_account_id, to_account_id, amount, type) VALUES ($1, $2, $3, $4) RETURNING id`
    err := r.db.QueryRow(query, tx.FromAccountID, tx.ToAccountID, tx.Amount, tx.Type).Scan(&id)
    return id, err
}

func (r *Repository) GetTransactionsByAccount(accountID int, from, to time.Time) ([]models.Transaction, error) {
    rows, err := r.db.Query(`SELECT id, from_account_id, to_account_id, amount, type, created_at FROM transactions 
        WHERE (from_account_id=$1 OR to_account_id=$1) AND created_at BETWEEN $2 AND $3`, accountID, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var txns []models.Transaction
    for rows.Next() {
        var t models.Transaction
        if err := rows.Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Type, &t.CreatedAt); err != nil {
            return nil, err
        }
        txns = append(txns, t)
    }
    return txns, nil
}

func (r *Repository) CreateCard(card *models.Card) (int, error) {
    var id int
    query := `INSERT INTO cards (account_id, encrypted_pan, encrypted_expiry, cvv_hash, pan_hmac) VALUES ($1,$2,$3,$4,$5) RETURNING id`
    err := r.db.QueryRow(query, card.AccountID, card.EncryptedPan, card.EncryptedExpiry, card.CvvHash, card.PanHMAC).Scan(&id)
    return id, err
}

func (r *Repository) GetCardsByAccountID(accountID int) ([]models.Card, error) {
    rows, err := r.db.Query(`SELECT id, account_id, encrypted_pan, encrypted_expiry, cvv_hash, pan_hmac, created_at FROM cards WHERE account_id=$1`, accountID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var cards []models.Card
    for rows.Next() {
        var c models.Card
        if err := rows.Scan(&c.ID, &c.AccountID, &c.EncryptedPan, &c.EncryptedExpiry, &c.CvvHash, &c.PanHMAC, &c.CreatedAt); err != nil {
            return nil, err
        }
        cards = append(cards, c)
    }
    return cards, nil
}

func (r *Repository) CreateCredit(credit *models.Credit) (int, error) {
    var id int
    query := `INSERT INTO credits (account_id, amount, interest_rate, monthly_payment, remaining_debt, start_date, end_date, status) 
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`
    err := r.db.QueryRow(query, credit.AccountID, credit.Amount, credit.InterestRate, credit.MonthlyPayment, credit.RemainingDebt,
        credit.StartDate, credit.EndDate, credit.Status).Scan(&id)
    return id, err
}

func (r *Repository) GetCreditByID(id int) (*models.Credit, error) {
    var c models.Credit
    query := `SELECT id, account_id, amount, interest_rate, monthly_payment, remaining_debt, start_date, end_date, status FROM credits WHERE id=$1`
    err := r.db.QueryRow(query, id).Scan(&c.ID, &c.AccountID, &c.Amount, &c.InterestRate, &c.MonthlyPayment, &c.RemainingDebt, &c.StartDate, &c.EndDate, &c.Status)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &c, err
}

func (r *Repository) UpdateCreditStatus(id int, status string) error {
    _, err := r.db.Exec(`UPDATE credits SET status=$1 WHERE id=$2`, status, id)
    return err
}

func (r *Repository) UpdateRemainingDebt(id int, newDebt float64) error {
    _, err := r.db.Exec(`UPDATE credits SET remaining_debt=$1 WHERE id=$2`, newDebt, id)
    return err
}

func (r *Repository) GetActiveCredits() ([]models.Credit, error) {
    rows, err := r.db.Query(`SELECT id, account_id, amount, interest_rate, monthly_payment, remaining_debt, start_date, end_date, status FROM credits WHERE status='active'`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var credits []models.Credit
    for rows.Next() {
        var c models.Credit
        if err := rows.Scan(&c.ID, &c.AccountID, &c.Amount, &c.InterestRate, &c.MonthlyPayment, &c.RemainingDebt, &c.StartDate, &c.EndDate, &c.Status); err != nil {
            return nil, err
        }
        credits = append(credits, c)
    }
    return credits, nil
}

func (r *Repository) CreatePaymentSchedule(p *models.PaymentSchedule) (int, error) {
    var id int
    query := `INSERT INTO payment_schedules (credit_id, due_date, payment_amount, paid) VALUES ($1,$2,$3,$4) RETURNING id`
    err := r.db.QueryRow(query, p.CreditID, p.DueDate, p.PaymentAmount, p.Paid).Scan(&id)
    return id, err
}

func (r *Repository) GetSchedulesByCreditID(creditID int) ([]models.PaymentSchedule, error) {
    rows, err := r.db.Query(`SELECT id, credit_id, due_date, payment_amount, paid, paid_at FROM payment_schedules WHERE credit_id=$1 ORDER BY due_date`, creditID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var schedules []models.PaymentSchedule
    for rows.Next() {
        var s models.PaymentSchedule
        if err := rows.Scan(&s.ID, &s.CreditID, &s.DueDate, &s.PaymentAmount, &s.Paid, &s.PaidAt); err != nil {
            return nil, err
        }
        schedules = append(schedules, s)
    }
    return schedules, nil
}

func (r *Repository) GetUnpaidSchedulesDueBefore(date time.Time) ([]models.PaymentSchedule, error) {
    rows, err := r.db.Query(`SELECT id, credit_id, due_date, payment_amount, paid FROM payment_schedules WHERE paid=false AND due_date < $1`, date)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var schedules []models.PaymentSchedule
    for rows.Next() {
        var s models.PaymentSchedule
        if err := rows.Scan(&s.ID, &s.CreditID, &s.DueDate, &s.PaymentAmount, &s.Paid); err != nil {
            return nil, err
        }
        schedules = append(schedules, s)
    }
    return schedules, nil
}

func (r *Repository) MarkSchedulePaid(id int, paidAt time.Time) error {
    _, err := r.db.Exec(`UPDATE payment_schedules SET paid=true, paid_at=$1 WHERE id=$2`, paidAt, id)
    return err
}
