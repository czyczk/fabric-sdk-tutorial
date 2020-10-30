package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type urlMethodPair struct {
	urlSuffix, method string
}

// EndpointMap is a map containing endpoints and the corresponding handlers that are defined and managed by a controller.
//
// Each entry in the map is organized in the following manner.
//   (urlSuffix, method): handler_function_list
// Thus it takes a URL suffix and an HTTP method as the key to perform a lookup.
type EndpointMap map[urlMethodPair][]gin.HandlerFunc

// A Controller must contain an endpoint map.
type Controller interface {
	GetGroupName() string
	GetEndpointMap() EndpointMap
}

// RegisterHandlers registers the endpoint handlers in the controller to the router group.
func RegisterHandlers(r *gin.RouterGroup, c Controller) error {
	group := r.Group(c.GetGroupName())

	em := c.GetEndpointMap()

	for pair, handlers := range em {
		if strings.EqualFold(pair.method, http.MethodGet) {
			group.GET(pair.urlSuffix, handlers...)
		} else if strings.EqualFold(pair.method, http.MethodPost) {
			group.POST(pair.urlSuffix, handlers...)
		} else if strings.EqualFold(pair.method, http.MethodPut) {
			group.PUT(pair.urlSuffix, handlers...)
		} else if strings.EqualFold(pair.method, http.MethodDelete) {
			group.DELETE(pair.urlSuffix, handlers...)
		} else if strings.EqualFold(pair.method, http.MethodPatch) {
			group.PATCH(pair.urlSuffix, handlers...)
		} else if strings.EqualFold(pair.method, http.MethodHead) {
			group.HEAD(pair.urlSuffix, handlers...)
		} else if strings.EqualFold(pair.method, http.MethodOptions) {
			group.OPTIONS(pair.urlSuffix, handlers...)
		} else {
			return fmt.Errorf("unsupported HTTP method")
		}
	}

	return nil
}
