package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "bank-api/middleware"
    "bank-api/service"
    "github.com/gorilla/mux"
)

type Handler struct {
    service *service.BankService
}

func NewHandler(s *service.BankService) *Handler {
    return &Handler{service: s}
}

type registerReq struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    user, err := h.service.Register(req.Username, req.Email, req.Password)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id":       user.ID,
        "username": user.Username,
        "email":    user.Email,
    })
}

type loginReq struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    user, err := h.service.ValidateCredentials(req.Email, req.Password)
    if err != nil {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }
    token, err := middleware.GenerateJWT(strconv.Itoa(user.ID))
    if err != nil {
        http.Error(w, "token generation failed", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    uid, _ := strconv.Atoi(userID)
    acc, err := h.service.CreateAccount(uid)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(acc)
}

type depositReq struct {
    Amount float64 `json:"amount"`
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    accID, _ := strconv.Atoi(vars["accountId"])
    var req depositReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    if err := h.service.Deposit(accID, req.Amount); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type transferReq struct {
    ToAccountId int     `json:"to_account_id"`
    Amount      float64 `json:"amount"`
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    uid, _ := strconv.Atoi(userID)
    var req transferReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    fromAccID, _ := strconv.Atoi(mux.Vars(r)["fromAccountId"])
    acc, err := h.service.GetAccountByID(fromAccID)
    if err != nil || acc.UserID != uid {
        http.Error(w, "access denied", http.StatusForbidden)
        return
    }
    if err := h.service.Transfer(fromAccID, req.ToAccountId, req.Amount); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type cardReq struct {
    CVV string `json:"cvv"`
}

func (h *Handler) CreateCard(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    uid, _ := strconv.Atoi(userID)
    vars := mux.Vars(r)
    accID, _ := strconv.Atoi(vars["accountId"])
    acc, err := h.service.GetAccountByID(accID)
    if err != nil || acc.UserID != uid {
        http.Error(w, "access denied", http.StatusForbidden)
        return
    }
    var req cardReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    card, pan, err := h.service.GenerateCard(accID, req.CVV)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(map[string]interface{}{
        "card_id": card.ID,
        "pan":     pan,
        "expiry":  "encrypted",
    })
}

func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    uid, _ := strconv.Atoi(userID)
    accID, _ := strconv.Atoi(mux.Vars(r)["accountId"])
    cards, err := h.service.GetCards(accID, uid)
    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }
    json.NewEncoder(w).Encode(cards)
}

type creditReq struct {
    Amount float64 `json:"amount"`
    Months int     `json:"months"`
}

func (h *Handler) CreateCredit(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    uid, _ := strconv.Atoi(userID)
    accID, _ := strconv.Atoi(mux.Vars(r)["accountId"])
    acc, err := h.service.GetAccountByID(accID)
    if err != nil || acc.UserID != uid {
        http.Error(w, "access denied", http.StatusForbidden)
        return
    }
    var req creditReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    credit, schedules, err := h.service.CreateCredit(accID, req.Amount, req.Months)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    json.NewEncoder(w).Encode(map[string]interface{}{
        "credit":   credit,
        "schedule": schedules,
    })
}

func (h *Handler) GetCreditSchedule(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    uid, _ := strconv.Atoi(userID)
    accID, _ := strconv.Atoi(r.URL.Query().Get("account_id"))
    year, _ := strconv.Atoi(r.URL.Query().Get("year"))
    month, _ := strconv.Atoi(r.URL.Query().Get("month"))
    income, expense, err := h.service.GetMonthlyAnalytics(accID, uid, year, month)
    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }
    json.NewEncoder(w).Encode(map[string]float64{"income": income, "expense": expense})
}

func (h *Handler) PredictBalance(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    uid, _ := strconv.Atoi(userID)
    accID, _ := strconv.Atoi(mux.Vars(r)["accountId"])
    days, _ := strconv.Atoi(r.URL.Query().Get("days"))
    if days == 0 {
        days = 30
    }
    pred, err := h.service.PredictBalance(accID, uid, days)
    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }
    json.NewEncoder(w).Encode(pred)
}
