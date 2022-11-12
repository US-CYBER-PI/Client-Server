package models

import "time"

type Token struct {
	Id          int
	Status      bool
	ExpiredDate time.Time
	Token       string
}
