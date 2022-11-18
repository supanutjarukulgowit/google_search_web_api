package service

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/supanutjarukulgowit/google_search_web_api/database"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	// DBg *gorm.DB
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

func (h *UserService) SignUp(req *model.SignUpRequest) error {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return err
	}
	password, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	id, _ := util.GetUUID()
	user := &model.User{
		Id:          id,
		Username:    req.Username,
		Password:    string(password),
		CreatedDate: time.Now(),
	}
	_ = user
	db.Create(user)

	return nil
}

func (h *UserService) SignIn(req *model.SignInRequest) (*http.Cookie, string, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return nil, "", err
	}
	user := &model.User{}
	db.Where("username = ?", req.Username).First(&user)
	if user.Id == "" {
		return nil, static.USER_AUTHEN_ERROR, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, static.USER_WRONG_PASSWORD, nil
	}
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.Id,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 1 day
	})

	token, err := claims.SignedString([]byte(static.SecretKey))
	if err != nil {
		return nil, "", err
	}
	cookie := new(http.Cookie)
	cookie.Name = "jwt"
	cookie.Value = token
	cookie.Expires = time.Now().Add(24 * time.Hour)
	return cookie, "", nil
}

func (h *UserService) User(cookie *http.Cookie) (*model.User, string, error) {
	db, err := h.Pg.ConnectPostgreSQLGorm(h.PgConnection.Host, h.PgConnection.User, h.PgConnection.Password, h.PgConnection.Database, h.PgConnection.Port)
	if err != nil {
		return nil, "", err
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(static.SecretKey), nil
	})
	if err != nil {
		return nil, static.USER_AUTHEN_ERROR, nil
	}
	claims := token.Claims.(*jwt.StandardClaims)
	user := &model.User{}
	db.Where("id = ?", claims.Issuer).First(user)
	return user, "", nil
}
