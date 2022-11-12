package repositories

import (
	"ClientServer/models"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryPG struct {
	db           *sql.DB
	queryUserRow string
	queryCheck   string
	queryToken   string
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
		db:           db,
		queryUserRow: "insert into users(phone,password,role_id) values ($1,$2,1);",
		queryCheck:   "select id from users WHERE phone = $1;",
		queryToken:   "select id,status,expired_date,token from tokens WHERE id = $1",
	}, nil
}

func (r *UserRepositoryPG) GetTokenById(id int) (*models.Token, error) {

	var token models.Token

	err := r.db.QueryRow(r.queryToken, id).Scan(&token.Id, &token.Status, &token.ExpiredDate, &token.Token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *UserRepositoryPG) CheckOccupancyPhone(phone string) bool {

	var user models.User

	err := r.db.QueryRow(r.queryCheck, phone).Scan(&user.Id)
	if err != nil {
		return false
	}

	return true
}

func (r *UserRepositoryPG) UserRegistration(phone, passwords string) bool {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwords), bcrypt.DefaultCost)
	if err != nil {
		return false
	}
	r.db.QueryRow(r.queryUserRow, phone, hashedPassword)
	return true
}
