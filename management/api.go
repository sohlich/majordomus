package management

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func withUserToken(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token")
		claims := map[string]interface{}(token.(*jwt.Token).Claims.(jwt.MapClaims))
		fmt.Println(claims["user"])
		newReq := r.WithContext(context.WithValue(r.Context(), "user", claims["user"]))
		h.ServeHTTP(w, newReq)
	}
}

func (m *mgmtModule) ApiAuthModule() http.Handler {
	mux := mux.NewRouter()
	mux.HandleFunc("/register", m.SignUp)
	mux.HandleFunc("/login", m.SignIn)
	return mux
}

func (m *mgmtModule) SignUp(w http.ResponseWriter, r *http.Request) {
	u := &AppUser{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(u)
	if err != nil {
		http.Error(w, "Cannot read user", 400)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.GetPassword()), 14)
	if err != nil {
		http.Error(w, "Cannot process password", 500)
	}
	u.Password = string(bytes)
	tx, err := m.Begin()
	if err != nil {
		http.Error(w, "Cannot register user", 500)
	}
	u.ID = uuid.NewV4().String()
	err = m.AddUser(tx, u)
	if err != nil {
		http.Error(w, "Cannot register user", 500)
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (m *mgmtModule) SignIn(w http.ResponseWriter, r *http.Request) {
	u := &AppUser{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(u)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Cannot read user", 400)
	}
	tx, err := m.Begin()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Cannot login user", 500)
	}
	dbUser, err := m.GetUserByEmail(tx, u.GetEmail())
	if err == sql.ErrNoRows {
		http.Error(w, "Bad login", 401)
	}
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Cannot login user", 500)
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.GetPassword()), []byte(u.GetPassword()))
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Bad login", 401)
	}
	jwt, err := generateJwtFrom(dbUser, m.JwtSecret)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "Authentication",
		Value:   "Bearer " + jwt,
		Domain:  m.Domain,
		Path:    "/",
		Expires: time.Now().Add(48 * time.Hour),
	})
}

func (m *mgmtModule) ApiGroupModule() http.Handler {
	mux := mux.NewRouter()
	mux.HandleFunc("/", m.ListGroups)
	return withUserToken(mux)
}

func (m *mgmtModule) ListGroups(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(string)

	tx, err := m.Begin()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
	}

}

func generateJwtFrom(u User, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": u.GetID(),
	})
	return token.SignedString([]byte(jwtSecret))
}
