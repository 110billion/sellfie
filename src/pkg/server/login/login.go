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
	"encoding/base64"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"io/ioutil"
	"math/rand"
	"net/http"
)

var (
	// store return cookie store
	store = sessions.NewCookieStore([]byte("secret"))
	log   logr.Logger
)

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
