package util

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

var inited bool = false

const CSRF_TOKEN_KEY = "csrf_token"

func getSessionObject(c echo.Context) (*sessions.Session, error) {

	var sess *sessions.Session
	var err error

	if sess, err = session.Get("session", c); err != nil {
		return nil, err
	}

	if !inited {
		c.Echo().Logger.Debug("initialize session")
		inited = true
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}
	}

	return sess, nil
}

func GetSession(c echo.Context, key string) (string, error) {

	if sess, err := getSessionObject(c); err != nil {
		return "", err
	} else {
		return sess.Values[key].(string), nil
	}
}

func SetSession(c echo.Context, key string, val string) error {

	if sess, err := getSessionObject(c); err != nil {
		return err
	} else {
		sess.Values[key] = val
		sess.Save(c.Request(), c.Response())
		return nil
	}
}
