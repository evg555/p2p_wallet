//nolint:revive
package handler

import (
	"net/http"

	"p2p_wallet/internal/api"
)

type handler struct{}

var _ api.ServerInterface = (*handler)(nil)

func New() *handler {
	return &handler{}
}

func (h *handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handler) GetUserById(w http.ResponseWriter, r *http.Request, id api.UserId) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *handler) LogoutUser(w http.ResponseWriter, r *http.Request, id api.UserId) {
	w.WriteHeader(http.StatusNotImplemented)
}
