package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"line_test/util"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo"
)

type AuthResponse struct {
	Code             string `form:"code"`
	State            string `form:"state"`
	Error            string `form:"error"`
	ErrorDescription string `form:"error_description"`
}

type TokenResp struct {
	AccessToken string `json:"access_token"`
}

func Token() echo.HandlerFunc {
	return func(c echo.Context) error {
		authResp := new(AuthResponse)
		if err := c.Bind(authResp); err != nil {
			return err
		}
		if authResp.Error != "" {
			return fmt.Errorf("error [%s]\nerror description [%s]", authResp.Error, authResp.ErrorDescription)
		}
		// CSRFチェックw
		if authResp.State != "1234567890" {
			return errors.New("csrf token check error")
		}

		clientId, err := util.GetSession(c, "client_id")
		if err != nil {
			return err
		}

		clientSecret, err := util.GetSession(c, "client_secret")
		if err != nil {
			return err
		}

		// token endpoint
		values := url.Values{}
		values.Set("grant_type", "authorization_code")
		values.Set("code", authResp.Code)
		values.Set("redirect_url", "https://sakashin.net/line_test/token")
		values.Set("client_id", clientId)
		values.Set("client_secret", clientSecret)
		req, err := http.NewRequest(
			"POST",
			"https://notify-bot.line.me/oauth/token",
			strings.NewReader(values.Encode()),
		)
		if err != nil {
			return err
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var tokenResponse = new(TokenResp)
		err = json.Unmarshal(b, tokenResponse)
		if err != nil {
			return err
		}

		c.Echo().Logger.Debug(fmt.Sprintf("access token [%s]", tokenResponse.AccessToken))

		var file *os.File
		if file, err = os.Create("~/lien_notify_token.txt"); err != nil {
			return err
		}

		defer file.Close()
		file.Write([]byte(tokenResponse.AccessToken + "\n"))
		c.Echo().Logger.Debug("access token saved")

		return GetNotify()(c)
	}
}
