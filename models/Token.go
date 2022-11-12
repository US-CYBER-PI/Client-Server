package models

type Token struct {
	Id          int32
	Status      bool
	ExpiredDate string
	Token       string
}
