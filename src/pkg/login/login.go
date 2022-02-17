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
	"encoding/base64"
	"golang.org/x/oauth2"
	"log"
	"math/rand"
	"net/http"
)

// User is user info name & email
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// MainView handles main page
func MainView(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(session.Values["user"])
}

func getLoginURL(oauthConf *oauth2.Config, state string) string {
	return oauthConf.AuthCodeURL(state)
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
