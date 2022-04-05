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

package auth

import (
	"github.com/110billion/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/usermanagerservice/src/pkg/server/auth/facebook"
	"github.com/110billion/usermanagerservice/src/pkg/server/auth/google"
	"github.com/110billion/usermanagerservice/src/pkg/server/auth/signup"
	"github.com/go-logr/logr"
)

type handler struct {
	googleHandler   apiserver.APIHandler
	facebookHandler apiserver.APIHandler
	signUpHandler   apiserver.APIHandler
}

// NewHandler instantiates a new apis handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{}

	// auth
	authWrapper := wrapper.New("/auth", nil, nil)
	if err := parent.Add(authWrapper); err != nil {
		return nil, err
	}

	// /auth/signup
	signUpHandler, err := signup.NewHandler(authWrapper, logger)
	if err != nil {
		return nil, err
	}
	handler.signUpHandler = signUpHandler

	// /auth/google
	googleHandler, err := google.NewHandler(authWrapper, logger)
	if err != nil {
		return nil, err
	}
	handler.googleHandler = googleHandler

	// /auth/facebook
	facebookHandler, err := facebook.NewHandler(authWrapper, logger)
	if err != nil {
		return nil, err
	}
	handler.facebookHandler = facebookHandler

	return handler, nil
}
