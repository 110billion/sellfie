package signin

import (
	"database/sql"
	"encoding/json"
	"github.com/110billion/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/usermanagerservice/src/internal/utils"
	"github.com/110billion/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/usermanagerservice/src/pkg/server/database"
	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	"net/http"
)

type handler struct {
	log logr.Logger
}

type signInReqBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Id       string `json:"id"`
	Password string `json:"password"`
}

// NewHandler instantiates a new facebook api handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{log: logger}

	// /signin
	signInWrapper := wrapper.New("/signin", []string{http.MethodPost}, handler.signInHandler)
	if err := parent.Add(signInWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}

func (h *handler) signInHandler(w http.ResponseWriter, req *http.Request) {
	// Decode request body
	signInReq := &signInReqBody{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(signInReq); err != nil {
		h.log.Error(err, "signin error")
		_ = utils.RespondError(w, http.StatusBadRequest, "request body is not in json form or is malformed")
		return
	}

	if signInReq.Id == "" || signInReq.Email == "" || signInReq.Password == "" || signInReq.Name == "" {
		_ = utils.RespondError(w, http.StatusBadRequest, "request body is not in json form or is malformed")
		return
	}

	// Open DB
	db, err := database.Connect()
	if err != nil {
		h.log.Error(err, "signin error")
		_ = utils.RespondError(w, http.StatusBadRequest, "db connection error")
		return
	}
	defer db.Close()

	var email string
	if err = db.QueryRow("SELECT user_email FROM USER_TABLE WHERE user_email = $1", signInReq.Email).Scan(&email); err != sql.ErrNoRows {
		h.log.Error(err, "signin error")
		_ = utils.RespondError(w, http.StatusBadRequest, "already existing email")
		return
	}

	var id string
	if err = db.QueryRow("SELECT user_id FROM USER_TABLE WHERE user_id = $1", signInReq.Id).Scan(&id); err != sql.ErrNoRows {
		h.log.Error(err, "signin error")
		_ = utils.RespondError(w, http.StatusBadRequest, "already existing id")
		return
	}

	// Insert User
	result, err := db.Exec("INSERT INTO USER_TABLE VALUES($1, $2, $3, $4)", signInReq.Email, signInReq.Name, signInReq.Password, signInReq.Id)
	if err != nil {
		h.log.Error(err, "signin error")
		_ = utils.RespondError(w, http.StatusBadRequest, "user registration error")
		return
	}

	n, err := result.RowsAffected()
	if n == 1 {
		_ = utils.RespondJSON(w, struct{}{})
	}
}
