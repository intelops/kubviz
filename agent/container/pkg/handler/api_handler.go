package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/intelops/kubviz/agent/container/api"
	"github.com/intelops/kubviz/agent/container/pkg/clients"
)

type APIHandler struct {
	conn *clients.NATSContext
}

const (
	appJSONContentType = "application/json"
	contentType        = "Content-Type"
)

func NewAPIHandler(conn *clients.NATSContext) (*APIHandler, error) {
	return &APIHandler{
		conn: conn,
	}, nil
}

func (ah *APIHandler) BindRequest(mux *chi.Mux) {
	mux.Route("/", func(r chi.Router) {
		api.HandlerFromMux(ah, r)
	})
}

func (ah *APIHandler) GetApiDocs(w http.ResponseWriter, r *http.Request) {
	swagger, err := api.GetSwagger()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set(contentType, appJSONContentType)
	_ = json.NewEncoder(w).Encode(swagger)
}

func (ah *APIHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(contentType, appJSONContentType)
	w.WriteHeader(http.StatusOK)
}
