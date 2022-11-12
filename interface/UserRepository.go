package _interface

import (
	"ClientServer/models"
	"time"
)

type UserRepository interface {
	GetTokenById(id int) (*models.Token, error)

	CreateToken(userId int) (*models.Token, error)

	UpdateToken(expiredDate time.Time, token string, tokenId int) (*models.Token, error)

	CheckOccupancyPhone(phone string) bool

	UserRegistration(phone, passwords string) bool

	GetUserById(id int) *models.User

	GetPaymentByKey(key string) *models.Payment

	UpdatePaymentStatus(id int, status string) bool
}
