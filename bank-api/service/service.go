package service

import (
    "errors"
    "math"
    "time"

    "bank-api/models"
    "bank-api/repository"
    "bank-api/utils"
    "github.com/sirupsen/logrus"
)

type BankService struct {
    repo      *repository.Repository
    emailServ *EmailService
    cbrClient *CBRClient
    hmacKey   []byte
}

func NewBankService(repo *repository.Repository, emailServ *EmailService, cbrClient *CBRClient, hmacKey string) *BankService {
    return &BankService{
        repo:      repo,
        emailServ: emailServ,
        cbrClient: cbrClient,
        hmacKey:   []byte(hmacKey),
    }
}

func (s *BankService) Register(username, email, password string) (*models.User, error) {
    existing, _ := s.repo.GetUserByEmail(email)
    if existing != nil {
        return nil, errors.New("email already exists")
    }
    existing, _ = s.repo.GetUserByUsername(username)
    if existing != nil {
        return nil, errors.New("username already exists")
    }
    hash, err := utils.HashPassword(password)
    if err != nil {
        return nil, err
    }
    user := &models.User{
        Username:     username,
        Email:        email,
        PasswordHash: hash,
    }
    if err := user.Validate(); err != nil {
        return nil, err
    }
    id, err := s.repo.CreateUser(user)
    if err != nil {
        return nil, err
    }
    user.ID = id
    logrus.WithField("user_id", id).Info("User registered")
    return user, nil
}

func (s *BankService) ValidateCredentials(email, password string) (*models.User, error) {
    user, err := s.repo.GetUserByEmail(email)
    if err != nil || user == nil {
        return nil, errors.New("invalid credentials")
    }
    if !utils.CheckPasswordHash(password, user.PasswordHash) {
        return nil, errors.New("invalid credentials")
    }
    return user, nil
}

func (s *BankService) CreateAccount(userID int) (*models.Account, error) {
    acc := &models.Account{
        UserID:   userID,
        Balance:  0,
        Currency: "RUB",
    }
    id, err := s.repo.CreateAccount(acc)
    if err != nil {
        return nil, err
    }
    acc.ID = id
    return acc, nil
}

func (s *BankService) Deposit(accountID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be positive")
    }
    acc, err := s.repo.GetAccountByID(accountID)
    if err != nil || acc == nil {
        return errors.New("account not found")
    }
    newBalance := acc.Balance + amount
    if err := s.repo.UpdateAccountBalance(accountID, newBalance); err != nil {
        return err
    }
    tx := &models.Transaction{
        ToAccountID: &accountID,
        Amount:      amount,
        Type:        "deposit",
    }
    _, _ = s.repo.CreateTransaction(tx)
    logrus.Infof("Deposit %.2f to account %d", amount, accountID)
    return nil
}

func (s *BankService) Transfer(fromAccountID, toAccountID int, amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be positive")
    }
    fromAcc, err := s.repo.GetAccountByID(fromAccountID)
    if err != nil || fromAcc == nil {
        return errors.New("source account not found")
    }
    toAcc, err := s.repo.GetAccountByID(toAccountID)
    if err != nil || toAcc == nil {
        return errors.New("destination account not found")
    }
    if fromAcc.Balance < amount {
        return errors.New("insufficient funds")
    }
    newFromBalance := fromAcc.Balance - amount
    newToBalance := toAcc.Balance + amount
    tx, err := s.repo.BeginTx()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    if err := s.repo.UpdateAccountBalanceTx(tx, fromAccountID, newFromBalance); err != nil {
        return err
    }
    if err := s.repo.UpdateAccountBalanceTx(tx, toAccountID, newToBalance); err != nil {
        return err
    }
    trans := &models.Transaction{
        FromAccountID: &fromAccountID,
        ToAccountID:   &toAccountID,
        Amount:        amount,
        Type:          "transfer",
    }
    if _, err := s.repo.CreateTransactionTx(tx, trans); err != nil {
        return err
    }
    if err := tx.Commit(); err != nil {
        return err
    }
    logrus.Infof("Transfer %.2f from account %d to %d", amount, fromAccountID, toAccountID)
    return nil
}

func (s *BankService) GenerateCard(accountID int, cvv string) (*models.Card, string, error) {
    acc, err := s.repo.GetAccountByID(accountID)
    if err != nil || acc == nil {
        return nil, "", errors.New("account not found")
    }
    pan := utils.GenerateLuhnNumber("400000", 16)
    if !utils.ValidateLuhn(pan) {
        return nil, "", errors.New("generated invalid pan")
    }
    expiry := time.Now().AddDate(5, 0, 0).Format("01/06")
    encPan, err := utils.EncryptPGP(pan)
    if err != nil {
        return nil, "", err
    }
    encExpiry, err := utils.EncryptPGP(expiry)
    if err != nil {
        return nil, "", err
    }
    cvvHash, err := utils.HashCVV(cvv)
    if err != nil {
        return nil, "", err
    }
    panHMAC := utils.ComputeHMAC(pan, s.hmacKey)
    card := &models.Card{
        AccountID:      accountID,
        EncryptedPan:   encPan,
        EncryptedExpiry: encExpiry,
        CvvHash:        cvvHash,
        PanHMAC:        panHMAC,
    }
    id, err := s.repo.CreateCard(card)
    if err != nil {
        return nil, "", err
    }
    card.ID = id
    return card, pan, nil
}

func (s *BankService) GetCards(accountID, userID int) ([]map[string]interface{}, error) {
    acc, err := s.repo.GetAccountByID(accountID)
    if err != nil || acc == nil || acc.UserID != userID {
        return nil, errors.New("access denied")
    }
    cards, err := s.repo.GetCardsByAccountID(accountID)
    if err != nil {
        return nil, err
    }
    result := []map[string]interface{}{}
    for _, c := range cards {
        pan, err := utils.DecryptPGP(c.EncryptedPan)
        if err != nil {
            logrus.WithError(err).Warn("Failed to decrypt PAN")
            continue
        }
        expiry, err := utils.DecryptPGP(c.EncryptedExpiry)
        if err != nil {
            continue
        }
        if !utils.VerifyHMAC(pan, c.PanHMAC, s.hmacKey) {
            logrus.Warn("HMAC mismatch for card")
            continue
        }
        result = append(result, map[string]interface{}{
            "id":      c.ID,
            "pan":     pan,
            "expiry":  expiry,
            "created": c.CreatedAt,
        })
    }
    return result, nil
}

func (s *BankService) GetAccountByID(id int) (*models.Account, error) {
    return s.repo.GetAccountByID(id)
}

func (s *BankService) CreateCredit(accountID int, amount float64, months int) (*models.Credit, []models.PaymentSchedule, error) {
    acc, err := s.repo.GetAccountByID(accountID)
    if err != nil || acc == nil {
        return nil, nil, errors.New("account not found")
    }
    rate, err := s.cbrClient.GetKeyRate()
    if err != nil {
        rate = 20.0
    }
    monthlyRate := rate / 100 / 12
    if monthlyRate < 0 {
        monthlyRate = 0.01
    }
    annuity := monthlyRate * math.Pow(1+monthlyRate, float64(months)) / (math.Pow(1+monthlyRate, float64(months)) - 1)
    monthlyPayment := amount * annuity
    remainingDebt := amount
    startDate := time.Now()
    endDate := startDate.AddDate(0, months, 0)

    credit := &models.Credit{
        AccountID:      accountID,
        Amount:         amount,
        InterestRate:   rate,
        MonthlyPayment: monthlyPayment,
        RemainingDebt:  remainingDebt,
        StartDate:      startDate,
        EndDate:        endDate,
        Status:         "active",
    }
    creditID, err := s.repo.CreateCredit(credit)
    if err != nil {
        return nil, nil, err
    }
    credit.ID = creditID

    schedules := []models.PaymentSchedule{}
    dueDate := startDate.AddDate(0, 1, 0)
    for i := 0; i < months; i++ {
        sch := models.PaymentSchedule{
            CreditID:      creditID,
            DueDate:       dueDate,
            PaymentAmount: monthlyPayment,
            Paid:          false,
        }
        id, err := s.repo.CreatePaymentSchedule(&sch)
        if err != nil {
            return nil, nil, err
        }
        sch.ID = id
        schedules = append(schedules, sch)
        dueDate = dueDate.AddDate(0, 1, 0)
    }
    logrus.Infof("Credit created for account %d, amount %.2f, months %d", accountID, amount, months)
    return credit, schedules, nil
}

func (s *BankService) GetMonthlyAnalytics(accountID, userID int, year, month int) (income, expense float64, err error) {
    acc, err := s.repo.GetAccountByID(accountID)
    if err != nil || acc == nil || acc.UserID != userID {
        return 0, 0, errors.New("access denied")
    }
    start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    end := start.AddDate(0, 1, 0)
    txns, err := s.repo.GetTransactionsByAccount(accountID, start, end)
    if err != nil {
        return 0, 0, err
    }
    for _, t := range txns {
        if t.Type == "deposit" || (t.ToAccountID != nil && *t.ToAccountID == accountID && t.Type == "transfer") {
            income += t.Amount
        } else if t.Type == "withdrawal" || (t.FromAccountID != nil && *t.FromAccountID == accountID) {
            expense += t.Amount
        }
    }
    return income, expense, nil
}

func (s *BankService) PredictBalance(accountID, userID int, days int) ([]float64, error) {
    if days > 365 {
        return nil, errors.New("max 365 days")
    }
    acc, err := s.repo.GetAccountByID(accountID)
    if err != nil || acc == nil || acc.UserID != userID {
        return nil, errors.New("access denied")
    }
    credits, _ := s.repo.GetActiveCredits()
    futurePayments := make(map[int]float64)
    now := time.Now()
    for _, credit := range credits {
        if credit.AccountID != accountID {
            continue
        }
        schedules, _ := s.repo.GetSchedulesByCreditID(credit.ID)
        for _, sch := range schedules {
            if sch.Paid {
                continue
            }
            daysDiff := int(sch.DueDate.Sub(now).Hours() / 24)
            if daysDiff >= 0 && daysDiff < days {
                futurePayments[daysDiff] += sch.PaymentAmount
            }
        }
    }
    prediction := make([]float64, days)
    balance := acc.Balance
    for d := 0; d < days; d++ {
        if payment, ok := futurePayments[d]; ok {
            balance -= payment
        }
        prediction[d] = balance
    }
    return prediction, nil
}
