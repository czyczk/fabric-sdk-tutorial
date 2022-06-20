package controller

import (
	"net/http"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// A ResourceController contains a group name and a `ResourceService` instance. It also implements the interface `Controller`.
type ResourceController struct {
	GroupName   string
	ResourceSvc service.ResourceServiceInterface
}

// GetGroupName returns the group name.
func (c *ResourceController) GetGroupName() string {
	return c.GroupName
}

// GetEndpointMap implements part of the interface `Controller`. It returns the API endpoints and handlers which are defined and managed by ResourceController.
func (c *ResourceController) GetEndpointMap() EndpointMap {
	return EndpointMap{
		urlMethodPair{"resource/:id/metadata", "GET"}: []gin.HandlerFunc{c.handleGetResourceMetadata},
	}
}

func (c *ResourceController) handleGetResourceMetadata(ctx *gin.Context) {
	id := ctx.Param("id")

	// Validity check
	pel := &ParameterErrorList{}
	id = pel.AppendIfEmptyOrBlankSpaces(id, "资源 ID 不能为空。")

	// Early return if there's parameter error
	if len(*pel) != 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, *pel)
		return
	}

	resDataMetadata, err := c.ResourceSvc.GetResourceMetadata(id)
	if err == nil {
		ctx.JSON(http.StatusOK, resDataMetadata)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		ctx.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}
