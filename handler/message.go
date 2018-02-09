package handler

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/line/line-bot-sdk-go/linebot"
)

func GetMessage() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "message", nil)
	}
}

func PostMessage() echo.HandlerFunc {

	type prm struct {
		Id      string `form:"id"`
		Message string `form:"message"`
	}

	// type msgJson struct {
	// 	To       string   `json:"to"`
	// 	Messages []string `json:"messages"`
	// }
	return func(c echo.Context) error {
		prm := new(prm)
		if err := c.Bind(prm); err != nil {
			return err
		}

		loginInfo := GetLoginInfo()
		client := &http.Client{}
		bot, err := linebot.New(loginInfo.Secret, loginInfo.ChannelAccessToken, linebot.WithHTTPClient(client))
		if err != nil {
			return err
		}
		if _, err = bot.PushMessage(prm.Id, linebot.NewTextMessage(prm.Message)).Do(); err != nil {
			return err
		}
		// msgJson := new(msgJson)
		// msgJson.To = prm.Id
		// msgJson.Messages = []string{prm.Message}
		// json, _ := json.Marshal(msgJson)

		// req, err := http.NewRequest(
		// 	"POST",
		// 	"https://api.line.me/v2/bot/message/push",
		// 	strings.NewReader(string(json)),
		// )
		// if err != nil {
		// 	return err
		// }

		// req.Header.Set("Content-Type", "application/json")
		// req.Header.Set("Authorization", "Bearer "+GetLoginInfo().ChannelAccessToken)

		// client := &http.Client{}
		// resp, err := client.Do(req)
		// if err != nil {
		// 	return err
		// }
		// defer resp.Body.Close()

		// b, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	return err
		// }

		// c.Echo().Logger.Debug("messaage response=" + string(b))

		return c.Render(http.StatusOK, "message_complete", nil)
	}
}
