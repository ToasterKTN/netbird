//go:build !openapivalidation

package apiv1alpha1

import "net/http"

// ValidationCompiled reports whether OpenAPI request validation is compiled in.
const ValidationCompiled = false

// V1ValidatorMiddleware is a placeholder in builds without openapivalidation.
type V1ValidatorMiddleware struct{}

// CreateV1ApiValidatingMiddleware always fails: validation is not compiled in.
func CreateV1ApiValidatingMiddleware() (*V1ValidatorMiddleware, error) {
	return nil, ErrValidationNotCompiled
}

// Handler is never reached because the constructor fails; it passes through.
func (v *V1ValidatorMiddleware) Handler(h http.Handler) http.Handler { return h }
