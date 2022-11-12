package main

import (
	_interface "ClientServer/interface"
	"ClientServer/repositories"
	_ "database/sql"
	"encoding/json"
	"fmt"
	qiwiSdk "github.com/US-CYBER-PI/qiwi-bill-paymentsgo-sdk/src"
	"github.com/US-CYBER-PI/qiwi-bill-paymentsgo-sdk/src/Models"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	pgUser         = "secret"
	pgPassword     = "secret"
	pgHost         = "localhost"
	pgPort         = "5432"
	pgDB           = "jwt"
	pgUserTable    = "users"
	pgLoginField   = "login"
	pgRoleTable    = "roles"
	pgRoleIdField  = "role_id"
	userRepository _interface.UserRepository
	qiwiToken      = "9f69ff96-d505-4ed1-84e2-f678867a5c23"
	qiwiSiteId     = "sa3khn-02"
	qiwiClient     qiwiSdk.Client
	hmacSecret     = []byte("c4bd7d88edb4fa1817abb11707958924384f7933e5facfd707dc1d1429af9936")
	port           = 9096
)

func init() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Error loading .env file")
	}

	if os.Getenv("PG_USER") != "" {
		pgUser = os.Getenv("PG_USER")
	}

	if os.Getenv("PG_PASSWORD") != "" {
		pgPassword = os.Getenv("PG_PASSWORD")
	}

	if os.Getenv("PG_HOST") != "" {
		pgHost = os.Getenv("PG_HOST")
	}

	if os.Getenv("PG_PORT") != "" {
		pgPort = os.Getenv("PG_PORT")
	}

	if os.Getenv("PG_DB") != "" {
		pgDB = os.Getenv("PG_DB")
	}

	if os.Getenv("PORT") != "" {
		port, _ = strconv.Atoi(os.Getenv("PORT"))
	}

	if os.Getenv("HMAC_SECRET") != "" {
		hmacSecret = []byte(os.Getenv("HMAC_SECRET"))
	}

	if os.Getenv("QIWI_TOKEN") != "" {
		qiwiToken = os.Getenv("QIWI_TOKEN")
	}

	if os.Getenv("QIWI_SITE_ID") != "" {
		qiwiSiteId = os.Getenv("QIWI_SITE_ID")
	}

}

func main() {

	var err error

	qiwiClient = *qiwiSdk.NewClient(qiwiToken, "https://api.qiwi.com/partner", qiwiSiteId)

	userRepository, err = repositories.NewUserRepositoryPG(pgHost, pgPort, pgUser, pgPassword, pgDB)

	if err != nil {
		panic(err)
	}

	http.HandleFunc("/api/v1/auth/reg", RegUser)
	http.HandleFunc("/api/v1/user/pay_token", PayToken)
	//http.HandleFunc("/api/v1/user/pay_token/status", PayTokenStatus)
	http.HandleFunc("/api/v1/user/phone/sms", PhoneSms)
	//http.HandleFunc("/api/v1/user/pay", Pay)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func RegUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	phone := r.FormValue("phone")
	password := r.FormValue("password")

	if phone == "" || password == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if !userRepository.CheckOccupancyPhone(phone) {
		result := userRepository.UserRegistration(phone, password)
		if !result {
			http.Error(w, "", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"mes": "successfully",
		})
		return
	}

	http.Error(w, "", http.StatusUnprocessableEntity)
}

func PayToken(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	id, _ := strconv.Atoi(r.Header.Get("User-Id"))

	user := userRepository.GetUserById(id)
	if user == nil {
		http.Error(w, "", http.StatusBadGateway)
		return
	}

	originalToken, _ := userRepository.GetTokenById(user.TokenId)

	if !originalToken.Status {
		token, err := userRepository.CreateToken(id)
		if err != nil {
			http.Error(w, "", http.StatusBadGateway)
			return
		}

		qiwiClient.GeneratePayToken(Models.GeneratePayToken{
			RequestId: strconv.Itoa(token.Id),
			Phone:     user.Phone,
			AccountId: strconv.Itoa(user.Id),
		})

		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"message": "success"}`, http.StatusTooEarly)
		return
	}

	//var tokenR, err = userRepository.GetTokenById(id)
	//
	//if err != nil {
	//	return nil
	//}

	//var datebd string
	//t, _ := time.Parse(tokenR.ExpiredDate, datebd)
	//t2 := time.Now()
	//dur := t2.Sub(t)
	//if dur*time.Hour > 24 {
	//	return err
	//}
	//
	exp := time.Now().Add(5 * time.Minute).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"key": fmt.Sprintf("%d|%d", time.Now().Unix(), id),
		"exp": exp,
	})
	tokenString, _ := token.SignedString(hmacSecret)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token":       tokenString,
		"expiredDate": exp,
	})
}

func PayTokenStatus(w http.ResponseWriter, r *http.Request) {

	//if r.Method != http.MethodPost {
	//	http.Error(w, "", http.StatusMethodNotAllowed)
	//	return
	//}
	//
	//id := r.FormValue("id")
	//amount := r.FormValue("amount")
	//status := r.FormValue("status")
	//
	//if id == "" || amount == "" || status == "" {
	//	http.Error(w, "", http.StatusBadRequest)
	//	return
	//}
	//var re Repository.UserRepositoryPG
	//
	//_, err := re.UserRegistration(id, token, expresion)
	//if err != nil {
	//	http.Error(w, "", http.StatusForbidden)
	//	return
	//}
}

func PhoneSms(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	code := r.FormValue("code")

	if code == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(r.Header.Get("User-Id"))

	user := userRepository.GetUserById(id)
	if user == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	resp := qiwiClient.SmsConfirm(Models.SmsConfirm{
		RequestId: strconv.Itoa(user.TokenId),
		SmsCode:   code,
	})

	f, err := json.Marshal(resp)
	if resp.Status.Value != "CREATED" {
		http.Error(w, string(f), http.StatusUnprocessableEntity)
		return
	}

	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, resp.TokenValue.TokenExpiredDate[:19])

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	_, err = userRepository.UpdateToken(t, resp.TokenValue.TokenValue, user.TokenId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"mes": "successfully",
	})
	return
}

func Pay(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/json")
	//
	//_ = json.NewEncoder(w).Encode(map[string]interface{}{
	//	"message": tokenString,
	//})
	//http.Error(w, "", http.StatusMethodNotAllowed)
	//return
}

func getToken(r *http.Request) (string, int) {

	if r.Method != http.MethodPost {
		return "", http.StatusMethodNotAllowed
	}

	token := r.Header.Get("Authorization")

	if token == "" {
		return "", http.StatusUnauthorized
	}
	extractedToken := strings.Split(token, "Bearer ")

	if len(extractedToken) < 2 {
		return "", http.StatusUnauthorized
	}

	return extractedToken[1], http.StatusOK
}
