package referenses

import "qiwi-clients/Models"

type UserRepository interface {
	Authentication(phone, password string, role_id, token_id int) (*Models.User, error)
}

type TokenRepository interface {
	AddToken(token string, seconds uint32) bool
	IsSet(token string) bool
}
