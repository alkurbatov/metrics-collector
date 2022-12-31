package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

func buildResponse(code int, msg string) string {
	return fmt.Sprintf("%d %s", code, msg)
}

func OK() string {
	return buildResponse(http.StatusOK, http.StatusText(http.StatusOK))
}

func loadViewTemplate(src string) *template.Template {
	view, err := template.ParseFiles(src)
	if err != nil {
		logging.Log.Panic(err)
	}

	return view
}
