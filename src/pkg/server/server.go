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

package server

import (
	"fmt"
	"github.com/110billion/usermanagerservice/src/pkg/login"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	logger = logf.Log.WithName("server")
)

const (
	port = 3550
)

// Server is an interface of server
type Server interface {
	Start()
}

// server is HTTP server for login API
type server struct {
	router *mux.Router
}

func New() *server {
	login.InitGoogleOauthConfig()

	r := mux.NewRouter()
	r.HandleFunc("/", login.MainView)
	r.HandleFunc("/auth/google/login", login.GoogleLoginHandler)
	r.HandleFunc("/auth/google/callback", login.GoogleAuthCallback)

	return &server{
		router: r,
	}
}

func (s *server) Start() {
	httpAddr := fmt.Sprintf("0.0.0.0:%d", port)

	logger.Info(fmt.Sprintf("Server is running on %s", httpAddr))
	if err := http.ListenAndServe(httpAddr, s.router); err != nil {
		logger.Error(err, "cannot launch http server")
		os.Exit(1)
	}
}
