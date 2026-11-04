package v1alpha1

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/netbirdio/netbird/management/internals/controllers/network_map"
	"github.com/netbirdio/netbird/management/server/account"
	"github.com/netbirdio/netbird/shared/management/http/apiv1alpha1"

	"github.com/netbirdio/netbird/management/server/permissions"

	peersmodule "github.com/netbirdio/netbird/management/internals/modules/peers"
	"github.com/netbirdio/netbird/management/server/api/v1alpha1/peers"
	"github.com/netbirdio/netbird/management/server/api/v1alpha1/users"
)

func NewAPIV1Handler(
	ctx context.Context,
	router *mux.Router,
	accountManager account.Manager,
	peerManager peersmodule.Manager,
	networkMapController network_map.Controller,
	permissionsManager permissions.Manager) (http.Handler, error) {

	v1validatorMiddleware, err := apiv1alpha1.CreateV1ApiValidatingMiddleware()
	if errors.Is(err, apiv1alpha1.ErrValidationNotCompiled) {
		// The handlers rely on the OpenAPI validator for input checks, so
		// never serve them unvalidated: answer every request with 501.
		log.WithContext(ctx).Warnf("%v; /api/v1alpha1 endpoints are disabled", err)
		router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, err.Error(), http.StatusNotImplemented)
		}))
		return router, nil
	}
	if err != nil {
		return nil, err
	}

	router.Use(v1validatorMiddleware.Handler)

	peersHandler := peers.NewHandler(accountManager, peerManager, networkMapController, permissionsManager)
	_ = peersHandler.WithEndpointsForRouter(router)

	usersHandler := users.NewHandler(accountManager)
	return usersHandler.WithEndpointsForRouter(router), nil
}
