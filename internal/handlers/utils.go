package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/log"
)

func writeErrorResponse(ctx context.Context, w http.ResponseWriter, code int, err error) {
	log.Ctx(ctx).Error().Err(err).Msg("")

	resp := fmt.Sprintf("%d %v", code, err)
	http.Error(w, resp, code)
}

func loadViewTemplate(src string) *template.Template {
	view, err := template.ParseFiles(src)
	if err != nil {
		log.Panic().Err(err).Msg("")
	}

	return view
}
