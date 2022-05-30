package upload

import (
	"github.com/110billion/sellfie/postmanagerservice/src/internal/apiserver"
	"github.com/110billion/sellfie/postmanagerservice/src/internal/wrapper"
	"github.com/go-logr/logr"
)

// NewHandler instantiates a new apis handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	return nil, nil
}
