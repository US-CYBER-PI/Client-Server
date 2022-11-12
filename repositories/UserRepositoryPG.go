package repositories

import (
	"ClientServer/models"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserRepositoryPG struct {
	db               *sql.DB
	queryInsertUser  string
	queryUpdateUser  string
	queryUser        string
	queryCheck       string
	queryToken       string
	queryCreateToken string
	queryUpdateToken string
}

func NewUserRepositoryPG(host, port, user, password, dbname string) (*UserRepositoryPG, error) {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &UserRepositoryPG{
		db:               db,
		queryInsertUser:  "insert into users(phone,password,role_id) values ($1,$2,1) RETURNING id;",
		queryUpdateUser:  "UPDATE users SET token_id = $1 WHERE id = $2",
		queryUser:        "SELECT id, phone, token_id FROM users WHERE id = $1",
		queryCheck:       "select id from users WHERE phone = $1;",
		queryToken:       "select id,status,expired_date,token from tokens WHERE id = $1",
		queryCreateToken: "insert into tokens(status,expired_date,token) values (false, now(),'') RETURNING id;",
		queryUpdateToken: "UPDATE tokens SET status = true, expired_date=$1, token=$2  WHERE id = $3",
	}, nil
}

func (u *UserRepositoryPG) GetTokenById(id int) (*models.Token, error) {

	var token models.Token

	err := u.db.QueryRow(u.queryToken, id).Scan(&token.Id, &token.Status, &token.ExpiredDate, &token.Token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (u *UserRepositoryPG) CreateToken(userId int) (*models.Token, error) {

	var tokenId int

	err := u.db.QueryRow(u.queryCreateToken).Scan(&tokenId)

	if err != nil {
		return nil, err
	}

	_, err = u.db.Exec(u.queryUpdateUser, tokenId, userId)
	if err != nil {
		return nil, err
	}

	return &models.Token{Id: tokenId}, nil
}

func (u *UserRepositoryPG) UpdateToken(expiredDate time.Time, token string, tokenId int) (*models.Token, error) {

	_, err := u.db.Exec(u.queryUpdateToken, expiredDate, token, tokenId)

	if err != nil {
		return nil, err
	}

	return &models.Token{
		Id:          tokenId,
		ExpiredDate: expiredDate,
		Token:       token,
		Status:      true,
	}, nil

}

func (u *UserRepositoryPG) CheckOccupancyPhone(phone string) bool {

	var user models.User

	err := u.db.QueryRow(u.queryCheck, phone).Scan(&user.Id)
	if err != nil {
		return false
	}

	return true
}

func (u *UserRepositoryPG) UserRegistration(phone, passwords string) bool {

	var userId int

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwords), bcrypt.DefaultCost)
	if err != nil {
		return false
	}

	err = u.db.QueryRow(u.queryInsertUser, phone, hashedPassword).Scan(&userId)
	if err != nil {
		return false
	}

	_, err = u.CreateToken(userId)
	if err != nil {
		return false
	}

	return true
}

func (u *UserRepositoryPG) GetUserById(id int) *models.User {
	var user models.User
	err := u.db.QueryRow(u.queryUser, id).Scan(&user.Id, &user.Phone, &user.TokenId)
	if err != nil {
		return nil
	}
	return &user
}
