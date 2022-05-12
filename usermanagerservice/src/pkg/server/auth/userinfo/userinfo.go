package userinfo

import (
	"github.com/110billion/sellfie/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/sellfie/usermanagerservice/src/internal/utils"
	"github.com/110billion/sellfie/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/sellfie/usermanagerservice/src/pkg/database"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"net/http"
)

type handler struct {
	log logr.Logger
}

type userInfoRespBody struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"image_url"`
	Comment string `json:"comment"`
}

// NewHandler instantiates a new userInfo api handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{log: logger}

	// /userinfo
	userInfoWrapper := wrapper.New("/userinfo/{id}", []string{http.MethodGet}, handler.userInfoHandler)
	if err := parent.Add(userInfoWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}

func (h *handler) userInfoHandler(w http.ResponseWriter, req *http.Request) {
	// Decode request
	vars := mux.Vars(req)
	id := vars["id"]

	if id == "" {
		_ = utils.RespondError(w, http.StatusBadRequest, "id is undefined")
		return
	}

	// Open DB
	db, err := database.Connect()
	if err != nil {
		h.log.Error(err, "get userinfo error")
		_ = utils.RespondError(w, http.StatusBadRequest, "db connection error")
		return
	}
	defer db.Close()

	var url, name, comment string
	if err = db.QueryRow("SELECT profile_url, name, profile_comment FROM USER_INFO WHERE user_id = $1", id).Scan(&url, &name, &comment); err != nil {
		h.log.Error(err, "get userinfo error")
		_ = utils.RespondError(w, http.StatusBadRequest, "cannot get user info")
		return
	}

	_ = utils.RespondJSON(w, userInfoRespBody{
		Id:      id,
		Name:    name,
		URL:     url,
		Comment: comment,
	})
}
