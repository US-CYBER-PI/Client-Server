package _interface

import "ClientServer/models"

type UserRepository interface {
	GetTokenById(id int) (*models.Token, error)

	CheckOccupancyPhone(phone string) bool

	UserRegistration(phone, passwords string) bool
}
