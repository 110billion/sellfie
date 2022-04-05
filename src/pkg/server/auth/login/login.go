/*
 Copyright 2021 The 110 billion Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package login

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/110billion/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/usermanagerservice/src/internal/utils"
	"github.com/110billion/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/usermanagerservice/src/pkg/server/database"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-logr/logr"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type claims struct {
	userEmail string
	jwt.StandardClaims
}

type loginResponse struct {
	ok    bool
	token string
}

type handler struct {
	log logr.Logger
}

var (
	// store return cookie store
	store = sessions.NewCookieStore([]byte("secret"))
	log   logr.Logger
)

type logInReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User is user info name & email
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// getLoginURL returns login url
func getLoginURL(oauthConf *oauth2.Config, state string) string {
	return oauthConf.AuthCodeURL(state)
}

// randToken returns random string for token
func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Login handles redirection to login page
func Login(w http.ResponseWriter, r *http.Request, oauthConfig *oauth2.Config) {
	session, _ := store.Get(r, "session")
	session.Options = &sessions.Options{
		MaxAge: 300,
	}
	state := randToken()
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		log.Error(err, "")
		return
	}
	http.Redirect(w, r, getLoginURL(oauthConfig, state), http.StatusTemporaryRedirect)
}

// Callback handles login check and redirection to main page
func Callback(w http.ResponseWriter, r *http.Request, oauthConfig *oauth2.Config, apiEndpoint string) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Error(err, "")
		return
	}

	state := session.Values["state"]
	delete(session.Values, "state")
	_ = session.Save(r, w)
	if state != r.FormValue("state") {
		http.Error(w, "Invalid session state", http.StatusUnauthorized)
		return
	}

	token, err := oauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cli := oauthConfig.Client(context.Background(), token)
	userInfoResp, err := cli.Get(apiEndpoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer userInfoResp.Body.Close()
	userInfo, err := ioutil.ReadAll(userInfoResp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var authUser User
	if err := json.Unmarshal(userInfo, &authUser); err != nil {
		log.Error(err, "")
		return
	}

	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400,
	}
	session.Values["user"] = authUser.Email
	session.Values["username"] = authUser.Name
	_ = session.Save(r, w)

	http.Redirect(w, r, "https://heychangju.shop", http.StatusFound)
}

// NewHandler instantiates a new facebook api handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{log: logger}

	// /signup
	logInWrapper := wrapper.New("/login", []string{http.MethodPost}, handler.logInHandler)
	if err := parent.Add(logInWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}

func (h *handler) logInHandler(w http.ResponseWriter, req *http.Request) {
	// Decode request body
	logInReq := &logInReqBody{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(logInReq); err != nil {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "request body is not in json form or is malformed")
		return
	}

	if logInReq.Email == "" || logInReq.Password == "" {
		_ = utils.RespondError(w, http.StatusBadRequest, "request body is not in json form or is malformed")
		return
	}

	// Open DB
	db, err := database.Connect()
	if err != nil {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "db connection error")
		return
	}
	defer db.Close()

	var email string
	if err = db.QueryRow("SELECT user_email FROM USER_TABLE WHERE user_email = $1", logInReq.Email).Scan(&email); err == sql.ErrNoRows {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "email not registered")
		return
	}

	var password string
	if err = db.QueryRow("SELECT password FROM USER_TABLE WHERE user_email = $1", logInReq.Email).Scan(&password); err == sql.ErrNoRows {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "internal error")
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(logInReq.Password))
	if err != nil {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "password doesn't match")
		return
	}

	jwtToken, err := getJwtToken(logInReq.Email)
	if err != nil {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "jwt token error")
		return
	}

	data := loginResponse{
		ok:    true,
		token: jwtToken,
	}
	_ = utils.RespondJSON(w, data)
}

func getJwtToken(email string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Hour)
	jwtKey := os.Getenv("JWT_SECRET_KEY")
	claims := &claims{
		userEmail: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
