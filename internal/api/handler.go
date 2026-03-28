package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"wallet-service/internal/models"
	"wallet-service/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type WalletHandler struct {
	service *service.WalletService
}

func NewWalletHandler(service *service.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

func (h *WalletHandler) ProcessOperation(w http.ResponseWriter, r *http.Request) {
	var req models.OperationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	//Валидация входных данных UUID и типа операции
	if req.WalletID == uuid.Nil {
		sendError(w, http.StatusBadRequest, "invalid wallet ID")
		return
	}

	wallet, err := h.service.ProcessOperation(r.Context(), &req)
	if err != nil {
		//Проверяем тип ошибки для определения статуса ответа
		errMsg := err.Error()
		if strings.Contains(errMsg, "insufficient funds") {
			//if err.Error() == "inssufficient funds" {
			sendError(w, http.StatusBadRequest, errMsg)
			return
		}
		if strings.Contains(errMsg, "wallet not found") {
			sendError(w, http.StatusNotFound, errMsg)
			return
		}
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	sendJSON(w, http.StatusOK, wallet)
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletIDStr := vars["id"]

	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid wallet ID format")
		return
	}
	balance, err := h.service.GetBalance(r.Context(), walletID)
	if err != nil {
		if strings.Contains(err.Error(), "wallet not found") {
			//if err.Error() == "wallet not found"
			sendError(w, http.StatusNotFound, "wallet not found")
			return
		}
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	response := models.BalanceResponse{
		WalletID: walletID,
		Balance:  balance,
	}
	sendJSON(w, http.StatusOK, response)
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, models.ErrorResponse{Error: message})
}
