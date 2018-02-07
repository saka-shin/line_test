package handler

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo"
)

func GetNotify() echo.HandlerFunc {
	return func(c echo.Context) error {

		var tokens []string

		pwd, _ := os.Getwd()
		if file, err := os.Open(pwd + "/line_notify_token.txt"); err == nil {
			defer file.Close()
			r := bufio.NewReader(file)
			for {
				token, err := r.ReadString('\n')
				if err == io.EOF {
					break
				}
				tokens = append(tokens, token[0:len(token)-1])
			}
		}

		return c.Render(http.StatusOK, "notify", tokens)
	}
}

type PostNotifyParam struct {
	Message string   `form:"message"`
	Tokens  []string `form:"tokens[]"`
}

func PostNotify() echo.HandlerFunc {
	return func(c echo.Context) error {

		prm := new(PostNotifyParam)
		if err := c.Bind(prm); err != nil {
			return err
		}

		for _, token := range prm.Tokens {
			go func() error {
				values := url.Values{}
				values.Set("message", prm.Message)
				req, err := http.NewRequest(
					"POST",
					"https://notify-api.line.me/api/notify",
					strings.NewReader(values.Encode()),
				)
				if err != nil {
					return err
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Authorization", "Bearer "+token)
				client := &http.Client{}
				_, err = client.Do(req)
				return err
			}()
		}

		return GetNotify()(c)
	}
}
