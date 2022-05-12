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
	"database/sql"
	"encoding/json"
	"github.com/110billion/sellfie/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/sellfie/usermanagerservice/src/internal/utils"
	"github.com/110billion/sellfie/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/sellfie/usermanagerservice/src/pkg/database"
	"github.com/110billion/sellfie/usermanagerservice/src/pkg/server/auth/token"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// Response is common struct for responding login request
type Response struct {
	Ok    bool   `json:"ok"`
	ID    string `json:"id"`
	Token string `json:"token"`
}

type handler struct {
	log logr.Logger
}

type logInReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewHandler instantiates a new login api handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{log: logger}

	// /login
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

	var email, password, id string
	if err = db.QueryRow("SELECT user_email, password, user_id FROM USER_TABLE WHERE user_email = $1", logInReq.Email).Scan(&email, &password, &id); err == sql.ErrNoRows {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "email not registered")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(logInReq.Password))
	if err != nil {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "password doesn't match")
		return
	}

	jwtToken, err := token.GetJwtToken(logInReq.Email)
	if err != nil {
		h.log.Error(err, "login error")
		_ = utils.RespondError(w, http.StatusBadRequest, "jwt token error")
		return
	}

	_ = utils.RespondJSON(w, Response{Ok: true, Token: jwtToken, ID: id})
}
