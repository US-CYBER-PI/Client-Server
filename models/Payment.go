package models

type Payment struct {
	ID                 int
	Amount             float64
	UserID             int
	Status             string
	PayTokenIdentifier string
}
