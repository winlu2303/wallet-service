package service

import (
	"context"
	"fmt"

	"wallet-service/internal/models"
	"wallet-service/internal/repository"

	"github.com/google/uuid"
)

type WalletService struct {
	repo repository.WalletRepository
}

func NewWalletService(repo repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	wallet, err := s.repo.GetWallet(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet: %w", err)
	}
	return wallet.Balance, nil
}

func (s *WalletService) ProcessOperation(ctx context.Context, req *models.OperationRequest) (*models.Wallet, error) {
	// Валидация входных данных
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if req.OperationType != "DEPOSIT" && req.OperationType != "WITHDRAW" {
		return nil, fmt.Errorf("invalid operation type: %s", req.OperationType)
	}
	// Обрабатываем операцию через репозиторий, который уже реализует логику блокировки и обновления баланса
	wallet, err := s.repo.UpdateWalletBalance(ctx, req.WalletID, req.OperationType, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to process operation: %w", err)
	}
	return wallet, nil
}
