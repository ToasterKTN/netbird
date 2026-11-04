package apiv1alpha1

import "errors"

// ErrValidationNotCompiled is returned by CreateV1ApiValidatingMiddleware when
// the binary was built without the openapivalidation tag.
var ErrValidationNotCompiled = errors.New("API v1alpha1 request validation not compiled in, rebuild with -tags openapivalidation")
