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

func generateJwtFrom(u User, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": u.GetID(),
	})
	return token.SignedString([]byte(jwtSecret))
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
		return
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.GetPassword()), 14)
	if err != nil {
		http.Error(w, "Cannot process password", 500)
		return
	}
	u.Password = string(bytes)
	tx, err := m.Begin()
	if err != nil {
		http.Error(w, "Cannot register user", 500)
		return
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
		return
	}
	defer r.Body.Close()
	tx, err := m.Begin()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Cannot login user", 500)
		return
	}
	dbUser, err := m.GetUserByEmail(tx, u.GetEmail())
	if err == sql.ErrNoRows {
		http.Error(w, "Bad login", 401)
		return
	}
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Cannot login user", 500)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.GetPassword()), []byte(u.GetPassword()))
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Bad login", 401)
		return
	}
	jwt, err := generateJwtFrom(dbUser, m.JwtSecret)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			m.ListGroupsHandler(w, r)
		}
		if r.Method == http.MethodPost {
			m.AddGroupHandler(w, r)
		}
	})
	return withUserToken(mux)
}

func (m *mgmtModule) ListGroupsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(string)
	tx, err := m.Begin()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	groups, err := m.Groups().GetGroupsBy(tx, user)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
}

func (m *mgmtModule) AddGroupHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(string)
	group := &IOTGroup{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(group)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	defer r.Body.Close()

	tx, err := m.Begin()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	group.ID = uuid.NewV4().String()
	group.OwnerID = user
	_, err = m.Groups().AddGroup(tx, group)
	if err != nil {
		fmt.Println(err.Error())
		tx.Rollback()
		http.Error(w, err.Error(), 500)
		return
	}
	tx.Commit()
}

func (m *mgmtModule) ApiDeviceModule() http.Handler {
	mx := mux.NewRouter()
	mx.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			m.ListDevicesHandler(w, r)
		case http.MethodPost:
			m.AddDeviceHandler(w, r)
		}
	})
	return withUserToken(mx)
}

func (m *mgmtModule) ListDevicesHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(string)
	tx, err := m.Begin()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	dLst, err := m.Devices().GetAllByUserID(tx, user)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	err = json.NewEncoder(w).Encode(dLst)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
}

func (m *mgmtModule) AddDeviceHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(string)
	device := &IOTDevice{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(device)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	tx, err := m.Begin()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	g, _ := m.GetGroupBy(tx, user, device.GetGroup())
	if g == nil {
		fmt.Println("Group not found")
		http.Error(w, "Group not found", 404)
		return
	}

	if err := m.Devices().AddDevice(tx, device); err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	tx.Commit()
}
