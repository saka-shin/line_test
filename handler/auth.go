package handler

import (
	"line_test/util"
	"net/http"
	"net/url"
	"strings"

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
		values := url.Values{}
		values.Set("response_type", "code")
		values.Set("client_id", authParam.ClientId)
		values.Set("redirect_url", "https://sakashin.net/line_test/token")
		values.Set("scope", "notify")
		values.Set("state", "1234567890") // csrf対策。テストなので適当に。
		values.Set("response_mode", "form_post")
		req, err := http.NewRequest(
			"GET",
			"https://notify-bot.line.me/oauth/authorize",
			strings.NewReader(values.Encode()),
		)
		if err != nil {
			return err
		}

		client := &http.Client{}
		if _, err := client.Do(req); err != nil {
			return err
		}

		return nil
	}
}
