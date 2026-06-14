package main

import (
    "fmt"
    "log"
    "net/http"

    "bank-api/handler"
    "bank-api/middleware"
    "bank-api/repository"
    "bank-api/service"
    "github.com/gorilla/mux"
    "github.com/sirupsen/logrus"
)

func main() {
    cfg := LoadConfig()
    logrus.SetFormatter(&logrus.JSONFormatter{})
    logrus.SetLevel(logrus.InfoLevel)

    connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
    repo, err := repository.NewRepository(connStr)
    if err != nil {
        logrus.Fatalf("Failed to connect to DB: %v", err)
    }
    defer repo.Close()

    emailSvc := service.NewEmailService(cfg)
    cbrClient := service.NewCBRClient()
    bankSvc := service.NewBankService(repo, emailSvc, cbrClient, cfg.HMACSecret)

    middleware.SetJWTSecret(cfg.JWTSecret)

    scheduler := service.NewScheduler(repo, emailSvc)
    scheduler.Start()
    defer scheduler.Stop()

    h := handler.NewHandler(bankSvc)

    r := mux.NewRouter()
    r.HandleFunc("/register", h.Register).Methods("POST")
    r.HandleFunc("/login", h.Login).Methods("POST")

    api := r.PathPrefix("/api").Subrouter()
    api.Use(middleware.AuthMiddleware)

    api.HandleFunc("/accounts", h.CreateAccount).Methods("POST")
    api.HandleFunc("/accounts/{accountId}/deposit", h.Deposit).Methods("POST")
    api.HandleFunc("/transfer/{fromAccountId}", h.Transfer).Methods("POST")
    api.HandleFunc("/accounts/{accountId}/cards", h.CreateCard).Methods("POST")
    api.HandleFunc("/accounts/{accountId}/cards", h.GetCards).Methods("GET")
    api.HandleFunc("/accounts/{accountId}/credits", h.CreateCredit).Methods("POST")
    api.HandleFunc("/credits/{creditId}/schedule", h.GetCreditSchedule).Methods("GET")
    api.HandleFunc("/analytics", h.GetAnalytics).Methods("GET")
    api.HandleFunc("/accounts/{accountId}/predict", h.PredictBalance).Methods("GET")

    logrus.Info("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
