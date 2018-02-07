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
	Code             string `form:"code" query:"code"`
	State            string `form:"state" query:"state"`
	Error            string `form:"error" query:"error"`
	ErrorDescription string `form:"error_description" query:"error_description"`
}

func (resp *AuthResponse) String() string {
	return fmt.Sprintf("code=%s, state=%s, error=%s, error description=%s", resp.Code, resp.State, resp.Error, resp.ErrorDescription)
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
		c.Echo().Logger.Debug("client id=" + clientId)

		clientSecret, err := util.GetSession(c, "client_secret")
		if err != nil {
			return err
		}
		c.Echo().Logger.Debug("client secret=" + clientSecret)
		c.Echo().Logger.Debug("code=" + authResp.Code)

		// token endpoint
		values := url.Values{}
		values.Set("grant_type", "authorization_code")
		values.Set("code", authResp.Code)
		values.Set("redirect_uri", util.BASE_URL+"/line_test/token")
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

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		} else if resp.StatusCode == 400 {
			return errors.New(string(b))
		}

		c.Echo().Logger.Debug("resp body=" + string(b))

		var tokenResponse = new(TokenResp)
		err = json.Unmarshal(b, tokenResponse)
		if err != nil {
			return err
		}

		c.Echo().Logger.Debug(fmt.Sprintf("access token [%s]", tokenResponse.AccessToken))

		var file *os.File
		pwd, _ := os.Getwd()
		// if file, err = os.Create(pwd + "/line_notify_token.txt"); err != nil {
		if file, err = os.OpenFile(pwd+"/line_notify_token.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return err
		}

		defer file.Close()
		file.Write([]byte(tokenResponse.AccessToken + "\n"))
		c.Echo().Logger.Debug("access token saved")

		return GetNotify()(c)
	}
}
