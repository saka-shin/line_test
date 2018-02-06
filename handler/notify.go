package handler

import (
	"bufio"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

func GetNotify() echo.HandlerFunc {
	return func(c echo.Context) error {

		var tokens []string

		if file, err := os.Open("~/line_notify_token.txt"); err == nil {
			defer file.Close()
			r := bufio.NewReader(file)
			for {
				token, err := r.ReadString('\n')
				if err == io.EOF {
					break
				}
				tokens = append(tokens, token)
			}
		}

		return c.Render(http.StatusOK, "notify", tokens)
	}
}

func PostNotify() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "notify", nil)
	}
}
