package openapi

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"
)

//go:embed *.gohtml
var files embed.FS

func Init(file string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		const operation = "Http.Resource.OpenAPI.Init"

		tpl, err := template.ParseFS(files, file+".gohtml")
		if err != nil {
			slog.ErrorContext(req.Context(), fmt.Errorf("%s (%s) -> page not found: %w", operation, req.RequestURI, err).Error())
			rw.WriteHeader(http.StatusInternalServerError)

			return
		}

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		data := map[string]any{
			"appName": "go-api-template",
			"specURL": "/docs/v1/go-api-template/swagger/doc.json",
		}
		if err = tpl.Execute(rw, data); err != nil {
			slog.ErrorContext(req.Context(), fmt.Errorf("%s (%s) -> execute template: %w", operation, req.RequestURI, err).Error())

			return
		}
	}
}
