package service

import (
	"context"
	"testing"

	"wallet-service/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	getWalletFunc           func(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error)
	updateWalletBalanceFunc func(ctx context.Context, walletID uuid.UUID, operationType string, amount int64) (*models.Wallet, error)
}

func (m *mockRepository) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	if m.getWalletFunc != nil {
		return m.getWalletFunc(ctx, walletID)
	}
	return nil, nil
}

func (m *mockRepository) UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, operationType string, amount int64) (*models.Wallet, error) {
	if m.updateWalletBalanceFunc != nil {
		return m.updateWalletBalanceFunc(ctx, walletID, operationType, amount)
	}
	return nil, nil
}

func TestProcessOperation_Deposit(t *testing.T) {
	walletID := uuid.New()
	expectedWallet := &models.Wallet{
		ID:      walletID,
		Balance: 1000,
		Version: 1,
	}

	mock := &mockRepository{
		updateWalletBalanceFunc: func(ctx context.Context, id uuid.UUID, opType string, amount int64) (*models.Wallet, error) {
			assert.Equal(t, walletID, id)
			assert.Equal(t, "DEPOSIT", opType)
			assert.Equal(t, int64(1000), amount)
			return expectedWallet, nil
		},
	}

	service := NewWalletService(mock)

	req := &models.OperationRequest{
		WalletID:      walletID,
		OperationType: "DEPOSIT",
		Amount:        1000,
	}
	result, err := service.ProcessOperation(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedWallet, result)
}

func TestProcessOperation_Withdraw_InsufficientFunds(t *testing.T) {
	walletID := uuid.New()
	mockRepo := &mockRepository{
		updateWalletBalanceFunc: func(ctx context.Context, id uuid.UUID, opType string, amount int64) (*models.Wallet, error) {
			return nil, assert.AnError
		},
	}
	service := NewWalletService(mockRepo)
	req := &models.OperationRequest{
		WalletID:      walletID,
		OperationType: "WITHDRAW",
		Amount:        1000,
	}
	_, err := service.ProcessOperation(context.Background(), req)
	assert.Error(t, err)
}

func TestGetBalance(t *testing.T) {
	walletID := uuid.New()
	expectedBalance := int64(500)

	mock := &mockRepository{
		getWalletFunc: func(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
			assert.Equal(t, walletID, id)
			return &models.Wallet{
				ID:      walletID,
				Balance: expectedBalance,
			}, nil
		},
	}
	service := NewWalletService(mock)

	balance, err := service.GetBalance(context.Background(), walletID)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
}

func TestProcessOperation_InvalidAmount(t *testing.T) {
	mock := &mockRepository{}
	service := NewWalletService(mock)

	req := &models.OperationRequest{
		WalletID:      uuid.New(),
		OperationType: "DEPOSIT",
		Amount:        -100,
	}
	_, err := service.ProcessOperation(context.Background(), req)
	assert.Error(t, err)
	// Проверяем точное сообщение об ошибке
	assert.Contains(t, err.Error(), "amount must be greater than zero")
}

func TestProcessOperation_InvalidOperationType(t *testing.T) {
	mock := &mockRepository{}
	service := NewWalletService(mock)

	req := &models.OperationRequest{
		WalletID:      uuid.New(),
		OperationType: "INVALID",
		Amount:        100,
	}
	_, err := service.ProcessOperation(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operation type")
}
