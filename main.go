package main

import (
	"io"
	"line_test/handler"

	"github.com/golang/go/src/pkg/html/template"
	"github.com/labstack/echo"
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

	// 全てのリクエストで差し込みたいミドルウェア（ログとか）はここ
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// テンプレート読み込み
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("template/*.html")),
	}

	// ルーティング
	e.GET("/", handler.Index())

	// サーバー起動
	e.Start(":1323") //ポート番号指定してね
}
