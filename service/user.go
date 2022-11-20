package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Pg           *database.PostgreSQL
	PgConnection *model.PostgreSQLConnect
}

func NewUserService(postgreSQL interface{}) (*UserService, error) {
	var pConnect model.PostgreSQLConnect
	err := util.InterfaceToStruct(postgreSQL, &pConnect)
	if err != nil {
		return nil, err
	}
	return &UserService{
		Pg:           database.NewPostgreSQL(),
		PgConnection: &pConnect,
	}, nil
}

func (h *UserService) SignUp(req *model.SignUpRequest) (string, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return "", fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	user := &model.User{}
	r := db.Where("username = ?", req.Username).First(&user)
	if r.Error != nil {
		return "", fmt.Errorf("find user error : %s", err.Error())
	}
	if user.Id == "" {
		return static.USER_ALREADY_SIGN_UP, nil
	}
	password, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	id, _ := util.GetUUID()
	newUser := &model.User{
		Id:          id,
		Username:    req.Username,
		Password:    string(password),
		CreatedDate: time.Now(),
	}
	r = db.Create(newUser)
	if r.Error != nil {
		return "", fmt.Errorf("create user error : %s", err.Error())
	}
	return "", nil
}

func (h *UserService) SignIn(req *model.SignInRequest) (string, string, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return "", "", fmt.Errorf("ConnectPostgreSQLGorm error : %s", err.Error())
	}
	user := &model.User{}
	r := db.Where("username = ?", req.Username).First(&user)
	if r.Error != nil {
		return "", "", fmt.Errorf("find username error : %s", err.Error())
	}
	if user.Id == "" {
		return "", static.USER_NOT_FOUND, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", static.USER_WRONG_PASSWORD, nil
	}
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.Id,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 1 day
	})

	token, err := claims.SignedString([]byte(static.SecretKey))
	if err != nil {
		return "", "", err
	}
	return token, "", nil
}
