package utils

import (
	"errors"

	"refina-transaction/config/env"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/types/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

func ConvertToResponseType(data interface{}) interface{} {
	return dto.TransactionsResponse{
		ID:              data.(model.Transactions).ID.String(),
		WalletID:        data.(model.Transactions).WalletID.String(),
		CategoryID:      data.(model.Transactions).CategoryID.String(),
		Amount:          data.(model.Transactions).Amount,
		TransactionDate: data.(model.Transactions).TransactionDate,
		Description:     data.(model.Transactions).Description,
	}
}

func VerifyToken(jwtToken string) (dto.UserData, error) {
	token, _ := jwt.Parse(jwtToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("parsing token error occured")
		}
		return []byte(env.Cfg.Server.JWTSecretKey), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return dto.UserData{}, errors.New("token is invalid")
	}

	return dto.UserData{
		ID:       claims["id"].(string),
		Username: claims["username"].(string),
		Email:    claims["email"].(string),
	}, nil
}

func ParseUUID(id string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsedID, nil
}
