package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"wallet-service/internal/models"

	"github.com/google/uuid"
)

type WalletRepository interface {
	GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error)
	UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, operationType string, amount int64) (*models.Wallet, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetWallet получает wallet по ID из database
func (r *PostgresRepository) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	query := `SELECT id, balance, version, created_at, updated_at 
	FROM wallets WHERE id = $1`
	var wallet models.Wallet
	err := r.db.QueryRowContext(ctx, query, walletID).Scan(&wallet.ID, &wallet.Balance, &wallet.Version, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return &wallet, nil
}

// UpdateWalletBalance обновляет баланс wallet в зависимости от типа операции (пополнение или списание) и возвращает обновленный wallet
// Если баланс недостаточен для списания, возвращает ошибку
func (r *PostgresRepository) UpdateWalletBalance(ctx context.Context, walletID uuid.UUID, operationType string, amount int64) (*models.Wallet, error) {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Блокируем строку wallet для обновления (SELECT FOR UPDATE)
	// Другие транзакции не смогут изменить эту строку, пока текущая транзакция не завершится
	// Это гарантирует, что мы получаем актуальный баланс и предотвращаем гонки данных
	query := `SELECT id, balance, version FROM wallets WHERE id = $1 FOR UPDATE`
	var wallet models.Wallet
	err = tx.QueryRowContext(ctx, query, walletID).Scan(&wallet.ID, &wallet.Balance, &wallet.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		return nil, fmt.Errorf("failed to lock wallet: %w", err)
	}

	// Вычисляем новый баланс в зависимости от типа операции
	newBalance := wallet.Balance
	switch operationType {
	case "DEPOSIT":
		newBalance += amount
	case "WITHDRAW":
		if wallet.Balance < amount {
			return nil, fmt.Errorf("insufficient funds: balance %d, withdrawal %d", wallet.Balance, amount)
		}
		newBalance -= amount
	default:
		return nil, fmt.Errorf("invalid operation type: %s", operationType)
	}

	// Обновляем баланс с проверкой версии wallet в базе данных
	updateQuery := `UPDATE wallets SET balance = $1, version = version + 1, updated_at = NOW() 
WHERE id = $2 AND version = $3
RETURNING id, balance, version, created_at, updated_at`

	err = tx.QueryRowContext(ctx, updateQuery, newBalance, walletID, wallet.Version).Scan(&wallet.ID, &wallet.Balance, &wallet.Version, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &wallet, nil
}
