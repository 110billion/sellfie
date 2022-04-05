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

package facebook

import (
	"github.com/110billion/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/usermanagerservice/src/pkg/server/auth/login"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"net/http"
	"os"
)

var (
	facebookOauthConfig *oauth2.Config
)

const (
	facebookRedirectURL         = "https://heychangju.shop/auth/facebook/callback"
	facebookUserInfoAPIEndpoint = "https://graph.facebook.com/me?fields=id,name,email"
)

type handler struct {
	log logr.Logger
}

// InitFacebookOauthConfig set facebook Oauth2 config when server starts
func InitFacebookOauthConfig() {
	facebookOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("FACEBOOK_ID"),
		ClientSecret: os.Getenv("FACEBOOK_SECRET"),
		RedirectURL:  facebookRedirectURL,
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
}

// NewHandler instantiates a new facebook api handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{log: logger}

	// /facebook
	facebookWrapper := wrapper.New("/facebook", nil, nil)
	if err := parent.Add(facebookWrapper); err != nil {
		return nil, err
	}

	// /facebook/login
	loginWrapper := wrapper.New("/login", nil, handler.loginHandler)
	if err := facebookWrapper.Add(loginWrapper); err != nil {
		return nil, err
	}

	// /facebook/callback
	callbackWrapper := wrapper.New("/callback", nil, handler.callbackHandler)
	if err := facebookWrapper.Add(callbackWrapper); err != nil {
		return nil, err
	}
	return handler, nil
}

// loginHandler handles redirection to facebook login
func (h *handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	login.Login(w, r, facebookOauthConfig)
}

// callbackHandler handles login check and redirection to main page
func (h *handler) callbackHandler(w http.ResponseWriter, r *http.Request) {
	login.Callback(w, r, facebookOauthConfig, facebookUserInfoAPIEndpoint)
}
