package response

import (
	"github.com/go-chi/render"
	"net/http"
)

type Response struct {
	Detail string `json:"detail" example:"response detail"`
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, err error, status int) {
	render.Status(r, status)
	render.JSON(w, r, Response{Detail: err.Error()})
}

func SuccessResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	render.Status(r, status)
	render.JSON(w, r, data)
}
