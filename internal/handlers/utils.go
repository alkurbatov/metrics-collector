package handlers

import (
	"fmt"
	"net/http"
)

func buildResponse(code int, msg string) string {
	return fmt.Sprintf("%d %s", code, msg)
}

func codeToResponse(code int) string {
	return buildResponse(code, http.StatusText(code))
}
