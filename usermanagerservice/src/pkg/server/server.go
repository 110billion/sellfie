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
	"github.com/110billion/sellfie/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/sellfie/usermanagerservice/src/internal/utils"
	"github.com/110billion/sellfie/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/sellfie/usermanagerservice/src/pkg/server/auth"
	"github.com/110billion/sellfie/usermanagerservice/src/pkg/server/auth/social/facebook"
	"github.com/110billion/sellfie/usermanagerservice/src/pkg/server/auth/social/google"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	log = logf.Log.WithName("user-manager-server")
)

const (
	port = 3550
)

// Server is an interface of server
type Server interface {
	Start()
}

// UserManagingServer is HTTP server for login API
type server struct {
	wrapper     wrapper.RouterWrapper
	authHandler apiserver.APIHandler
}

// New is a constructor of Server
func New() (Server, error) {
	google.InitGoogleOauthConfig()
	facebook.InitFacebookOauthConfig()

	srv := &server{}
	srv.wrapper = wrapper.New("/", nil, srv.rootHandler)

	srv.wrapper.SetRouter(mux.NewRouter())
	srv.wrapper.Router().HandleFunc("/", srv.rootHandler)

	// Set apisHandler
	authHandler, err := auth.NewHandler(srv.wrapper, log)
	if err != nil {
		return nil, err
	}
	srv.authHandler = authHandler

	return srv, nil
}

func (s *server) Start() {
	addr := fmt.Sprintf("0.0.0.0:%d", port)

	log.Info(fmt.Sprintf("Server is running on %s", addr))
	if err := http.ListenAndServe(addr, s.wrapper.Router()); err != nil { // TODO: TLS
		log.Error(err, "cannot launch http server")
		os.Exit(1)
	}
}

func (s *server) rootHandler(w http.ResponseWriter, _ *http.Request) {
	paths := metav1.RootPaths{}
	addPath(&paths.Paths, s.wrapper)

	_ = utils.RespondJSON(w, paths)
}

// addPath adds all the leaf API endpoints
func addPath(paths *[]string, w wrapper.RouterWrapper) {
	if w.Handler() != nil {
		*paths = append(*paths, w.FullPath())
	}

	for _, c := range w.Children() {
		addPath(paths, c)
	}
}
