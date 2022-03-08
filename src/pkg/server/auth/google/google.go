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

package google

import (
	"github.com/110billion/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/usermanagerservice/src/pkg/server/login"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	"os"
)

const (
	googleRedirectURL         = "heychangju.shop/auth/google/callback"
	googleUserInfoAPIEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"
	googleScopeEmail          = "https://www.googleapis.com/auth/userinfo.email"
	googleScopeProfile        = "https://www.googleapis.com/auth/userinfo.profile"
)

var (
	googleOauthConfig *oauth2.Config
)

type handler struct {
	log logr.Logger
}

// InitGoogleOauthConfig set google Oauth2 config when server starts
func InitGoogleOauthConfig() {
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_ID"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{googleScopeEmail, googleScopeProfile},
		Endpoint:     google.Endpoint,
	}
}

// NewHandler instantiates a new google api handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{log: logger}

	// /google
	googleWrapper := wrapper.New("/google", nil, nil)
	if err := parent.Add(googleWrapper); err != nil {
		return nil, err
	}

	// /google/login
	loginWrapper := wrapper.New("/login", nil, handler.loginHandler)
	if err := googleWrapper.Add(loginWrapper); err != nil {
		return nil, err
	}

	// /google/callback
	callbackWrapper := wrapper.New("/callback", nil, handler.callbackHandler)
	if err := googleWrapper.Add(callbackWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}

// loginHandler handles redirection to google login
func (h *handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	login.Login(w, r, googleOauthConfig)
}

// callbackHandler handles login check and redirection to main page
func (h *handler) callbackHandler(w http.ResponseWriter, r *http.Request) {
	login.Callback(w, r, googleOauthConfig, googleUserInfoAPIEndpoint)
}
