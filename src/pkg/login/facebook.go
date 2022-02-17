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
	"golang.org/x/oauth2/facebook"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	facebookOauthConfig *oauth2.Config
)

const (
	facebookRedirectURL         = "{domain}/auth/facebook/callback" // TODO: Pull From loadbalancer or etc.
	facebookUserInfoAPIEndpoint = "https://graph.facebook.com/me?fields=id,name,email"
)

// InitFacebookOauthConfig set facebook Oauth2 config when server starts
func InitFacebookOauthConfig() {
	facebookOauthConfig = &oauth2.Config{
		ClientID:     "", // TODO: Pull from secret
		ClientSecret: "",
		RedirectURL:  facebookRedirectURL,
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
}

// FBLoginHandler handles redirection to google login
func FBLoginHandler(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, getLoginURL(facebookOauthConfig, state), http.StatusTemporaryRedirect)
}

// FBAuthCallback handles redirection to google login
func FBAuthCallback(w http.ResponseWriter, r *http.Request) {
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

	token, err := facebookOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cli := facebookOauthConfig.Client(context.Background(), token)
	userInfoResp, err := cli.Get(facebookUserInfoAPIEndpoint)
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
