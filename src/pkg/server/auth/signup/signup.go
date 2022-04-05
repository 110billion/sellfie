package signup

import (
	"database/sql"
	"encoding/json"
	"github.com/110billion/usermanagerservice/src/internal/apiserver"
	"github.com/110billion/usermanagerservice/src/internal/utils"
	"github.com/110billion/usermanagerservice/src/internal/wrapper"
	"github.com/110billion/usermanagerservice/src/pkg/server/database"
	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type handler struct {
	log logr.Logger
}

type signUpReqBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Id       string `json:"id"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
}

// NewHandler instantiates a new facebook api handler
func NewHandler(parent wrapper.RouterWrapper, logger logr.Logger) (apiserver.APIHandler, error) {
	handler := &handler{log: logger}

	// /signup
	signUpWrapper := wrapper.New("/signup", []string{http.MethodPost}, handler.signUpHandler)
	if err := parent.Add(signUpWrapper); err != nil {
		return nil, err
	}

	return handler, nil
}

func (h *handler) signUpHandler(w http.ResponseWriter, req *http.Request) {
	// Decode request body
	signUpReq := &signUpReqBody{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(signUpReq); err != nil {
		h.log.Error(err, "signup error")
		_ = utils.RespondError(w, http.StatusBadRequest, "request body is not in json form or is malformed")
		return
	}

	if signUpReq.Id == "" || signUpReq.Email == "" || signUpReq.Password == "" || signUpReq.Name == "" {
		_ = utils.RespondError(w, http.StatusBadRequest, "request body is not in json form or is malformed")
		return
	}

	// Open DB
	db, err := database.Connect()
	if err != nil {
		h.log.Error(err, "signup error")
		_ = utils.RespondError(w, http.StatusBadRequest, "db connection error")
		return
	}
	defer db.Close()

	var email string
	if err = db.QueryRow("SELECT user_email FROM USER_TABLE WHERE user_email = $1", signUpReq.Email).Scan(&email); err != sql.ErrNoRows {
		h.log.Error(err, "signup error")
		_ = utils.RespondError(w, http.StatusBadRequest, "already existing email")
		return
	}

	var id string
	if err = db.QueryRow("SELECT user_id FROM USER_TABLE WHERE user_id = $1", signUpReq.Id).Scan(&id); err != sql.ErrNoRows {
		h.log.Error(err, "signup error")
		_ = utils.RespondError(w, http.StatusBadRequest, "already existing id")
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(signUpReq.Password), bcrypt.DefaultCost)

	// Insert User
	result, err := db.Exec("INSERT INTO USER_TABLE VALUES($1, $2, $3, $4)", signUpReq.Email, signUpReq.Name, password, signUpReq.Id)
	if err != nil {
		h.log.Error(err, "signup error")
		_ = utils.RespondError(w, http.StatusBadRequest, "user registration error")
		return
	}

	n, err := result.RowsAffected()
	if n == 1 {
		_ = utils.RespondJSON(w, struct{}{})
	}
}
