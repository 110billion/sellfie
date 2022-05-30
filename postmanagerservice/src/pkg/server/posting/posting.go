package posting

import (
	"github.com/110billion/sellfie/postmanagerservice/src/internal/apiserver"
	"github.com/110billion/sellfie/postmanagerservice/src/internal/wrapper"
	"github.com/110billion/sellfie/postmanagerservice/src/pkg/server/posting/delete"
	"github.com/110billion/sellfie/postmanagerservice/src/pkg/server/posting/upload"
	"github.com/go-logr/logr"
)

type handler struct {
	uploadHandler apiserver.APIHandler
	deleteHandler apiserver.APIHandler
}

// NewHandler instantiates a new apis handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{}

	// /posting
	postingWrapper := wrapper.New("/posting", nil, nil)
	if err := parent.Add(postingWrapper); err != nil {
		return nil, err
	}

	// /posting/upload
	uploadHandler, err := upload.NewHandler(postingWrapper, logger)
	if err != nil {
		return nil, err
	}
	handler.uploadHandler = uploadHandler

	// /posting/delete
	deleteHandler, err := delete.NewHandler(postingWrapper, logger)
	if err != nil {
		return nil, err
	}
	handler.deleteHandler = deleteHandler

	return handler, nil
}
