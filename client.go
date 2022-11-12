package main

import (
	_interface "ClientServer/interface"
	"ClientServer/repositories"
	"ClientServer/utils"
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
	"time"
)

var (
	pgUser         = "secret"
	pgPassword     = "secret"
	pgHost         = "localhost"
	pgPort         = "5432"
	pgDB           = "jwt"
	userRepository _interface.UserRepository
	qiwiToken      = "8f69ff16-d505-1ed1-84e3-f677467a5c23"
	qiwiSiteId     = "sa4kjn-12"
	qiwiClient     qiwiSdk.Client
	hmacSecret     = []byte("c3bd7d88edb4fa1817abb11702158924384f7933e5facfd707dc1d1429af9931")
	port           = 9096
	jwtManager     *utils.JwtManager
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

	jwtManager = utils.NewJwtManager(hmacSecret)

	qiwiClient = *qiwiSdk.NewClient(qiwiToken, "https://api.qiwi.com/partner", qiwiSiteId)

	userRepository, err = repositories.NewUserRepositoryPG(pgHost, pgPort, pgUser, pgPassword, pgDB)

	if err != nil {
		panic(err)
	}

	http.HandleFunc("/api/v1/auth/reg", RegUser)
	http.HandleFunc("/api/v1/user/pay_token", PayToken)
	http.HandleFunc("/api/v1/user/phone/sms", PhoneSms)
	http.HandleFunc("/api/v1/user/pay_token/status", PayStatus)
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

func PayStatus(w http.ResponseWriter, r *http.Request) {

	userPayToken := r.FormValue("user_pay_token")

	if userPayToken == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	claim := jwtManager.GetTokenClaim(userPayToken)

	if claim == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	payment := userRepository.GetPaymentByKey((*claim)["key"].(string))

	if payment == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if payment.Status != "COMPLETED" {
		resp := qiwiClient.GetPayment(strconv.Itoa(payment.ID))
		if payment.Status != resp.Status.Value {
			userRepository.UpdatePaymentStatus(payment.ID, resp.Status.Value)
			payment.Status = resp.Status.Value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     payment.ID,
		"amount": payment.Amount,
		"status": payment.Status,
	})
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
