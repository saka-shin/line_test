package handler

import (
	"fmt"
	"line_test/util"
	"net/http"

	"github.com/labstack/echo"
)

type AuthParam struct {
	ClientId     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
}

type AuthResp struct {
	Code             string `json:"code"`
	State            string `json:"state"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Auth
// LINE連携の認証をリクエストする。
// 認証結果を/line_test/tokenにコールバックする。
func Auth() echo.HandlerFunc {
	return func(c echo.Context) error {
		authParam := new(AuthParam)
		if err := c.Bind(authParam); err != nil {
			return err
		}

		// client_idとclient_secretをセッションに保存しておく
		if err := util.SetSession(c, "client_id", authParam.ClientId); err != nil {
			return err
		}
		if err := util.SetSession(c, "client_secret", authParam.ClientSecret); err != nil {
			return err
		}

		// authorization endpoint
		return c.Redirect(http.StatusMovedPermanently,
			fmt.Sprintf("https://notify-bot.line.me/oauth/authorize?response_type=code&client_id=%s&redirect_url=https://sakashin.net/line_test/token&scope=notify&state=1234567890&response_mode=form_post", authParam.ClientId))
	}
}
