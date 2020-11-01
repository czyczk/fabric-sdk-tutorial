package controller

import (
	"net/http"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"github.com/gin-gonic/gin"
)

// A PingPongController implements the interface `Controller`.
type PingPongController struct {
	GroupName string
	ScrewSvc  *service.ScrewService
}

// GetGroupName returns the group name
func (ppc *PingPongController) GetGroupName() string {
	return ppc.GroupName
}

// GetEndpointMap implements the interface `Controller` and returns the API endpoints and handlers defined and managed by ScrewController.
func (ppc *PingPongController) GetEndpointMap() EndpointMap {
	pingHandler := func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	}

	return EndpointMap{
		urlMethodPair{"/ping", "GET"}: []gin.HandlerFunc{
			pingHandler,
		},
		urlMethodPair{"/ping", "POST"}: []gin.HandlerFunc{
			pingHandler,
		},
	}
}
