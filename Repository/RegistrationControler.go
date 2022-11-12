package Repository

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"qiwi-clients/Models"
)

type UserRepositoryPG struct {
	db           *sql.DB
	queryUserRow string
	queryCheak   string
	queryToken   string
}

func ConnectionDB(host, port, user, password, dbname string) (*UserRepositoryPG, error) {

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
		queryUserRow: "insert into user(phone,passwords,role_id) values ($1,$2,1);",
		queryCheak:   "select id,phone,passwords,role_id,token_id from users WHERE phone = $1;",
		queryToken:   "select id,status,expired_date,token from tokens WHERE id = $1",
	}, nil
}

func (r *UserRepositoryPG) GetTokenbyId(id int) (*Models.Token, error) {

	var Token_ Models.Token

	err := r.db.QueryRow(r.queryToken, id).Scan(&Token_.Id, &Token_.Status, &Token_.Expired_date, &Token_.Token)
	if err != nil {
		return nil, err
	}
	return &Token_, nil
}

func (r *UserRepositoryPG) CheakUserRegistation(phone string) bool {

	err := r.db.QueryRow(r.queryUserRow, phone).Err()
	if err != nil {
		return false
	}

	return true
}

func (r *UserRepositoryPG) UserRegistration(phone, passwords string) (*Models.User, error) {

	var user Models.User

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwords), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(hashedPassword))
	r.db.QueryRow(r.queryUserRow, phone, hashedPassword)
	return &user, nil
}
