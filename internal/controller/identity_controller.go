package controller

import (
	"net/http"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type IdentityController struct {
	GroupName   string
	DocumentSvc service.DocumentServiceInterface
	AuthSvc     service.AuthServiceInterface
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
		urlMethodPair{"auths/request-list", "GET"}: []gin.HandlerFunc{c.handleGetAuthRequestList},
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
	// Extract and check parameters
	pageSizeStr := c.Query("pageSize")
	bookmark := processBase64FromURLQuery(c.Query("bookmark"))

	pel := &ParameterErrorList{}
	pageSize := 10
	if strings.TrimSpace(pageSizeStr) != "" {
		pageSize = pel.AppendIfNotPositiveInt(pageSizeStr, "分页大小应为正整数。")
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// ListDocumentIDsByCreator
	resourceIDs, err := ic.DocumentSvc.ListDocumentIDsByCreator(pageSize, bookmark)

	// Check error type and generate the corresponding response
	if err == nil {
		c.JSON(http.StatusOK, resourceIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (ic *IdentityController) handleGetAuthPendingList(c *gin.Context) {
	// Extract and check parameters
	pageSizeStr := c.Query("pageSize")
	bookmark := processBase64FromURLQuery(c.Query("bookmark"))

	pel := &ParameterErrorList{}
	pageSize := 10
	if strings.TrimSpace(pageSizeStr) != "" {
		pageSize = pel.AppendIfNotPositiveInt(pageSizeStr, "分页大小应为正整数。")
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// ListPendingAuthSessionIDsByResourceCreator
	authSessionIDs, err := ic.AuthSvc.ListPendingAuthSessionIDsByResourceCreator(pageSize, bookmark)

	// Check error type and generate the corresponding response
	if err == nil {
		c.JSON(http.StatusOK, authSessionIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (ic *IdentityController) handleGetAuthRequestList(c *gin.Context) {
	// Extract and check parameters
	pageSizeStr := c.Query("pageSize")
	bookmark := processBase64FromURLQuery(c.Query("bookmark"))
	isLatestFirstStr := c.Query("isLatestFirst")

	pel := &ParameterErrorList{}

	pageSize := 10
	if strings.TrimSpace(pageSizeStr) != "" {
		pageSize = pel.AppendIfNotPositiveInt(pageSizeStr, "分页大小应为正整数。")
	}

	isLatestFirst := true
	if strings.TrimSpace(isLatestFirstStr) != "" {
		isLatestFirst = pel.AppendIfNotBool(isLatestFirstStr, "置最新于最前选项应为 bool 值。")
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// ListAuthSessionIDsByRequestor
	authSessionIDs, err := ic.AuthSvc.ListAuthSessionIDsByRequestor(pageSize, bookmark, isLatestFirst)

	// Check error type and generate the corresponding response
	if err == nil {
		c.JSON(http.StatusOK, authSessionIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
