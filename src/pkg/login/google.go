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
	"encoding/json"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	RedirectURL         = "http://localhost:3550/auth/google/callback" // TODO: Pull From loadbalancer or etc.
	UserInfoAPIEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"
	ScopeEmail          = "https://www.googleapis.com/auth/userinfo.email"
	ScopeProfile        = "https://www.googleapis.com/auth/userinfo.profile"
)

var (
	googleOauthConfig *oauth2.Config
	store             = sessions.NewCookieStore([]byte("secret"))
)

func InitGoogleOauthConfig() {
	googleOauthConfig = &oauth2.Config{
		ClientID:     "", // TODO: Pull from secret
		ClientSecret: "",
		RedirectURL:  RedirectURL,
		Scopes:       []string{ScopeEmail, ScopeProfile},
		Endpoint:     google.Endpoint,
	}
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options = &sessions.Options{
		MaxAge: 300,
	}
	state := randToken()
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		log.Println(err.Error())
		return
	}
	http.Redirect(w, r, getLoginURL(googleOauthConfig, state), http.StatusTemporaryRedirect)
}

func GoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println(err.Error())
	}

	state := session.Values["state"]
	delete(session.Values, "state")
	_ = session.Save(r, w)
	if state != r.FormValue("state") {
		http.Error(w, "Invalid session state", http.StatusUnauthorized)
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cli := googleOauthConfig.Client(context.Background(), token)
	userInfoResp, err := cli.Get(UserInfoAPIEndpoint)
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
		log.Println(err.Error())
	}

	session.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400,
	}
	session.Values["user"] = authUser.Email
	session.Values["username"] = authUser.Name
	_ = session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}
