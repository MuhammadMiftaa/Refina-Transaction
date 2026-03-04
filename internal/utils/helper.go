package utils

import (
	"time"

	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"

	"github.com/google/uuid"
)

func ConvertToResponseType(data any) any {
	return dto.TransactionsResponse{
		ID:              data.(model.Transactions).ID.String(),
		WalletID:        data.(model.Transactions).WalletID.String(),
		CategoryID:      data.(model.Transactions).CategoryID.String(),
		CategoryName:    data.(model.Transactions).Category.Name,
		CategoryType:    string(data.(model.Transactions).Category.Type),
		Amount:          data.(model.Transactions).Amount,
		TransactionDate: data.(model.Transactions).TransactionDate,
		Description:     data.(model.Transactions).Description,
	}
}

func ParseUUID(id string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsedID, nil
}

func Ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}
