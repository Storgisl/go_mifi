package models

import (
    "errors"
    "regexp"
    "time"
)

type User struct {
    ID           int       `json:"id" db:"id"`
    Username     string    `json:"username" db:"username"`
    Email        string    `json:"email" db:"email"`
    PasswordHash string    `json:"-" db:"password_hash"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Account struct {
    ID        int       `json:"id" db:"id"`
    UserID    int       `json:"user_id" db:"user_id"`
    Balance   float64   `json:"balance" db:"balance"`
    Currency  string    `json:"currency" db:"currency"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Card struct {
    ID              int       `json:"id" db:"id"`
    AccountID       int       `json:"account_id" db:"account_id"`
    EncryptedPan    string    `json:"-" db:"encrypted_pan"`
    EncryptedExpiry string    `json:"-" db:"encrypted_expiry"`
    CvvHash         string    `json:"-" db:"cvv_hash"`
    PanHMAC         string    `json:"-" db:"pan_hmac"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type Transaction struct {
    ID            int       `json:"id" db:"id"`
    FromAccountID *int      `json:"from_account_id,omitempty" db:"from_account_id"`
    ToAccountID   *int      `json:"to_account_id,omitempty" db:"to_account_id"`
    Amount        float64   `json:"amount" db:"amount"`
    Type          string    `json:"type" db:"type"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type Credit struct {
    ID             int       `json:"id" db:"id"`
    AccountID      int       `json:"account_id" db:"account_id"`
    Amount         float64   `json:"amount" db:"amount"`
    InterestRate   float64   `json:"interest_rate" db:"interest_rate"`
    MonthlyPayment float64   `json:"monthly_payment" db:"monthly_payment"`
    RemainingDebt  float64   `json:"remaining_debt" db:"remaining_debt"`
    StartDate      time.Time `json:"start_date" db:"start_date"`
    EndDate        time.Time `json:"end_date" db:"end_date"`
    Status         string    `json:"status" db:"status"`
}

type PaymentSchedule struct {
    ID            int        `json:"id" db:"id"`
    CreditID      int        `json:"credit_id" db:"credit_id"`
    DueDate       time.Time  `json:"due_date" db:"due_date"`
    PaymentAmount float64    `json:"payment_amount" db:"payment_amount"`
    Paid          bool       `json:"paid" db:"paid"`
    PaidAt        *time.Time `json:"paid_at,omitempty" db:"paid_at"`
}

func (u *User) Validate() error {
    if u.Username == "" || len(u.Username) < 3 {
        return errors.New("username must be at least 3 characters")
    }
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
    if !emailRegex.MatchString(u.Email) {
        return errors.New("invalid email format")
    }
    return nil
}
