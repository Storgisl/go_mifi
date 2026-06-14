package service

import (
    "time"

    "bank-api/repository"
    "github.com/sirupsen/logrus"
)

type Scheduler struct {
    repo      *repository.Repository
    emailServ *EmailService
    stopCh    chan struct{}
}

func NewScheduler(repo *repository.Repository, emailServ *EmailService) *Scheduler {
    return &Scheduler{
        repo:      repo,
        emailServ: emailServ,
        stopCh:    make(chan struct{}),
    }
}

func (s *Scheduler) Start() {
    ticker := time.NewTicker(12 * time.Hour)
    go func() {
        for {
            select {
            case <-ticker.C:
                s.processOverduePayments()
            case <-s.stopCh:
                ticker.Stop()
                return
            }
        }
    }()
    logrus.Info("Scheduler started: checking overdue payments every 12 hours")
}

func (s *Scheduler) Stop() {
    close(s.stopCh)
    logrus.Info("Scheduler stopped")
}

func (s *Scheduler) processOverduePayments() {
    logrus.Info("Checking overdue payments...")
    now := time.Now()
    schedules, err := s.repo.GetUnpaidSchedulesDueBefore(now)
    if err != nil {
        logrus.WithError(err).Error("Failed to get overdue schedules")
        return
    }
    for _, sch := range schedules {
        credit, err := s.repo.GetCreditByID(sch.CreditID)
        if err != nil || credit == nil {
            continue
        }
        account, err := s.repo.GetAccountByID(credit.AccountID)
        if err != nil || account == nil {
            continue
        }
        penalty := sch.PaymentAmount * 0.1
        totalDue := sch.PaymentAmount + penalty
        if account.Balance >= totalDue {
            newBalance := account.Balance - totalDue
            if err := s.repo.UpdateAccountBalance(account.ID, newBalance); err != nil {
                logrus.WithError(err).Error("Failed to deduct penalty")
                continue
            }
            if err := s.repo.MarkSchedulePaid(sch.ID, time.Now()); err != nil {
                logrus.WithError(err).Error("Failed to mark schedule paid")
            }
            newDebt := credit.RemainingDebt - sch.PaymentAmount
            if newDebt < 0 {
                newDebt = 0
            }
            _ = s.repo.UpdateRemainingDebt(credit.ID, newDebt)
            if newDebt <= 0 {
                _ = s.repo.UpdateCreditStatus(credit.ID, "paid")
            }
            logrus.Infof("Auto-payment with penalty processed for credit %d, amount %.2f", credit.ID, totalDue)

            user, _ := s.repo.GetUserByID(account.UserID)
            if user != nil {
                _ = s.emailServ.SendPaymentNotification(user.Email, totalDue, "Платеж по кредиту (включая штраф)")
            }
        } else {
            logrus.Warnf("Insufficient funds for credit %d, penalty not applied", credit.ID)
        }
    }
}
