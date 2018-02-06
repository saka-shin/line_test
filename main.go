package main

import (
	"io"
	"line_test/handler"

	"github.com/gorilla/sessions"

	"github.com/golang/go/src/pkg/html/template"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Debug = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	// テンプレート読み込み
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("template/*.html")),
	}

	g := e.Group("/line_test")

	// 静的ファイル
	g.Static("/css", "static/css")

	// ルーティング
	g.GET("/", handler.Index())
	g.POST("/auth", handler.Auth())
	g.POST("/token", handler.Token())
	g.GET("/notify", handler.GetNotify())

	// サーバー起動
	e.Logger.Fatal(e.Start(":1323"))
}
