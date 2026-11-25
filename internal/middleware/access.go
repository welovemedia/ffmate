package middleware

import (
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"goyave.dev/goyave/v5"
)

// MetaAccess is the route meta key defining which access an api key requires for a route.
//
//	router.Get("/tasks", c.list).SetMeta(middleware.MetaAccess, middleware.AccessRead)
//	router.Post("/tasks", c.add).SetMeta(middleware.MetaAccess, middleware.AccessWrite)
//
// It only applies to requests authenticated by an api key. Requests authenticated by
// basic auth, or any request while basic auth is disabled, are always allowed.
const MetaAccess = "ffmate.access"

const (
	AccessRead  = dto.APIKeyAccessRead
	AccessWrite = dto.APIKeyAccessWrite
	// AccessBasicAuth restricts a route to basic auth, no api key is accepted
	AccessBasicAuth dto.APIKeyAccess = "basicAuth"
)

// hasAccess reports whether an api key with the given access may call the matched route.
// Routes without (valid) access meta allow every valid api key.
func hasAccess(route *goyave.Route, granted dto.APIKeyAccess) bool {
	meta, ok := route.LookupMeta(MetaAccess)
	if !ok {
		return true
	}

	required, ok := meta.(dto.APIKeyAccess)
	if !ok {
		return true
	}

	switch required {
	case AccessRead:
		return granted == AccessRead || granted == AccessWrite
	case AccessWrite:
		return granted == AccessWrite
	case AccessBasicAuth:
		return false
	default:
		return true
	}
}
