package interceptor

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/supanutjarukulgowit/google_search_web_api/static"
	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

func ValidateToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t := strings.Replace(c.Request().Header.Get("Authorization"), "Bearer ", "", 1)
			token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
				_, ok := token.Method.(*jwt.SigningMethodHMAC)
				if !ok {
					return nil, fmt.Errorf("unauthorized")
				}
				return []byte(static.SecretKey), nil
			})
			if err != nil {
				respErr := util.GenError(c, static.UN_AUTH_ERROR, "unauth error", static.UN_AUTH_ERROR, http.StatusUnauthorized)
				return c.JSON(http.StatusBadRequest, respErr)
			}
			if token.Valid {
				return next(c)
			} else {
				respErr := util.GenError(c, static.UN_AUTH_ERROR, "unauth error", static.UN_AUTH_ERROR, http.StatusUnauthorized)
				return c.JSON(http.StatusBadRequest, respErr)
			}
		}
	}
}
