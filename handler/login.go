package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"line_test/util"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo"
	yaml "gopkg.in/yaml.v2"
)

type LoginInfo struct {
	Id                 string `yaml:"id"`
	Secret             string `yaml:"secret"`
	ChannelAccessToken string `yaml:"channel_access_token"`
}

func GetLoginInfo() *LoginInfo {
	yml, err := ioutil.ReadFile("line_login.yml")
	if err != nil {
		panic(err)
	}

	loginInfo := new(LoginInfo)
	err = yaml.Unmarshal(yml, loginInfo)
	if err != nil {
		panic(err)
	}

	return loginInfo
}

func LoginIndex() echo.HandlerFunc {

	type viewItem struct {
		LoginInfo
		BaseURL string
	}

	return func(c echo.Context) error {
		vi := new(viewItem)
		vi.LoginInfo = *GetLoginInfo()
		vi.BaseURL = util.BASE_URL
		return c.Render(http.StatusOK, "login_index", vi)
	}
}

func LoginCallback() echo.HandlerFunc {

	type param struct {
		Code                    string `query:"code"`
		State                   string `query:"state"`
		FriendshipStatusChanged bool   `query:"friendship_status_changed"`
		Error                   string `query:"error"`
		ErrorDescription        string `query:"error_description"`
	}

	type tokenResp struct {
		AccessToken   string `json:"access_token"`
		ExpiresIn     int    `json:"expires_in"`
		IdToken       string `json:"id_token"`
		RefereshToken string `json:"refresh_token"`
		Scope         string `json:"scope"`
		TokenType     string `json:"token_type"`
	}

	return func(c echo.Context) error {

		prm := new(param)
		if err := c.Bind(prm); err != nil {
			return err
		} else if prm.Error != "" {
			return fmt.Errorf("error [%s]\nerror description [%s]", prm.Error, prm.ErrorDescription)
		}

		c.Echo().Logger.Debug(fmt.Sprintf("code=%s,state=%s,friendship_status_changed=%t,error=%s,error_description=%s", prm.Code, prm.State, prm.FriendshipStatusChanged, prm.Error, prm.ErrorDescription))

		values := url.Values{}
		values.Set("grant_type", "authorization_code")
		values.Set("code", prm.Code)
		values.Set("redirect_uri", util.BASE_URL+"/line_login/callback")

		loginInfo := GetLoginInfo()
		values.Set("client_id", loginInfo.Id)
		values.Set("client_secret", loginInfo.Secret)

		req, err := http.NewRequest(
			"POST",
			"https://api.line.me/oauth2/v2.1/token",
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

		c.Echo().Logger.Debug("token response=" + string(b))

		var tr = new(tokenResp)
		err = json.Unmarshal(b, tr)
		if err != nil {
			return err
		}

		idToken, err := decodeIdToken(tr.IdToken, c)
		if err != nil {
			return err
		}

		var file *os.File
		pwd, _ := os.Getwd()
		if file, err = os.OpenFile(pwd+"/line_login_token.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return err
		}

		defer file.Close()
		file.Write([]byte(idToken.Payload + "\n"))
		c.Echo().Logger.Debug("id token saved")

		return c.Render(http.StatusOK, "login_complete", nil)
	}
}

type idTokenPayload struct {
	Iss     string `json:"iss"`
	Sub     string `json:"sub"`
	Aud     string `json:"aud"`
	Exp     int64  `json:"exp"`
	Iat     int    `json:"iat"`
	Nonce   string `json:"nonce"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Payload string
}

func decodeIdToken(s string, c echo.Context) (idTokenPayload, error) {

	var pl = new(idTokenPayload)

	// step1
	ss := strings.Split(s, ".")

	// step2
	// header := base64Decode(ss[0])
	payload := base64Decode(ss[1])
	signature := hex.EncodeToString([]byte(base64Decode(ss[2])))

	// step3
	if sigCheck := checkSignature(GetLoginInfo().Secret, ss[0]+"."+ss[1], signature, c); !sigCheck {
		return *pl, errors.New("signature check failed")
	}

	if err := json.Unmarshal([]byte(payload), pl); err != nil {
		return *pl, err
	}

	// step4
	if pl.Iss != "https://access.line.me" {
		return *pl, errors.New("invalid iss")
	}

	// step5
	if pl.Aud != GetLoginInfo().Id {
		return *pl, errors.New("invalid aud")
	}

	// step6
	if time.Now().Unix() > pl.Exp {
		return *pl, errors.New("invalid exp")
	}

	pl.Payload = payload

	return *pl, nil
}

func base64Decode(str string) string {
	decoded, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return string(decoded)
}

func checkSignature(key string, target string, signature string, c echo.Context) bool {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(target))
	return hex.EncodeToString(mac.Sum(nil)) == signature
}
