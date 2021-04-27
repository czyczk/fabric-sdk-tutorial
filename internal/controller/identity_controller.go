package controller

import (
	"net/http"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
	userIdentity, err := ic.IdentitySvc.GetIdentityInfo()
	if err == nil {
		c.JSON(http.StatusOK, userIdentity)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (ic *IdentityController) handleGetDocumentList(c *gin.Context) {

}

func (ic *IdentityController) handleGetAuthPendingList(c *gin.Context) {

}
