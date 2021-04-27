package controller

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"github.com/gin-gonic/gin"
)

type IdentityController struct {
	GroupName   string
	IdentitySvc service.IdentityServiceInterface
}

// GetGroupName returns the group name.
func (c *IdentityController) GetGroupName() string {
	return c.GroupName
}

// GetEndpointMap implements part of the interface `Controller`. It returns the API endpoints and handlers which are defined and managed by DocumentController.
func (c *IdentityController) GetEndpointMap() EndpointMap {
	return EndpointMap{
		urlMethodPair{"", "GET"}:                   []gin.HandlerFunc{c.handleGetIdentity},
		urlMethodPair{"documents/list", "GET"}:     []gin.HandlerFunc{c.handleGetDocumentList},
		urlMethodPair{"auths/pending-list", "GET"}: []gin.HandlerFunc{c.handleGetAuthPendingList},
	}
}

func (ic *IdentityController) handleGetIdentity(c *gin.Context) {

}

func (ic *IdentityController) handleGetDocumentList(c *gin.Context) {

}

func (ic *IdentityController) handleGetAuthPendingList(c *gin.Context) {

}
