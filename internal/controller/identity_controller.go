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
	GroupName      string
	DocumentSvc    service.DocumentServiceInterface
	EntityAssetSvc service.EntityAssetServiceInterface
	AuthSvc        service.AuthServiceInterface
	IdentitySvc    service.IdentityServiceInterface
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
		urlMethodPair{"assets/list", "GET"}:        []gin.HandlerFunc{c.handleGetEntityAssetList},
		urlMethodPair{"auths/pending-list", "GET"}: []gin.HandlerFunc{c.handleGetAuthPendingList},
		urlMethodPair{"auths/request-list", "GET"}: []gin.HandlerFunc{c.handleGetAuthRequestList},
	}
}

func (c *IdentityController) handleGetIdentity(ctx *gin.Context) {
	userIdentity, err := c.IdentitySvc.GetIdentityInfo()
	if err == nil {
		ctx.JSON(http.StatusOK, userIdentity)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *IdentityController) handleGetDocumentList(ctx *gin.Context) {
	// Extract and check parameters
	pel := &ParameterErrorList{}
	isLatestFirst := true
	isLatestFirstStr := ctx.Query("isLatestFirst")
	if isLatestFirstStr != "" {
		isLatestFirst = pel.AppendIfNotBool(isLatestFirstStr, "最新于最前选项必须为 bool 值。")
	}
	pageSizeStr := ctx.Query("pageSize")
	bookmark := processBase64FromURLQuery(ctx.Query("bookmark"))

	pageSize := 10
	if strings.TrimSpace(pageSizeStr) != "" {
		pageSize = pel.AppendIfNotPositiveInt(pageSizeStr, "分页大小应为正整数。")
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// ListDocumentIDsByCreator
	resourceIDs, err := c.DocumentSvc.ListDocumentIDsByCreator(isLatestFirst, pageSize, bookmark)

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, resourceIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *IdentityController) handleGetEntityAssetList(ctx *gin.Context) {
	// Extract and check parameters
	pel := &ParameterErrorList{}
	isLatestFirst := true
	isLatestFirstStr := ctx.Query("isLatestFirst")
	if isLatestFirstStr != "" {
		isLatestFirst = pel.AppendIfNotBool(isLatestFirstStr, "最新于最前选项必须为 bool 值。")
	}
	pageSizeStr := ctx.Query("pageSize")
	bookmark := processBase64FromURLQuery(ctx.Query("bookmark"))

	pageSize := 10
	if strings.TrimSpace(pageSizeStr) != "" {
		pageSize = pel.AppendIfNotPositiveInt(pageSizeStr, "分页大小应为正整数。")
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// ListEntityAssetIDsByCreator
	resourceIDs, err := c.EntityAssetSvc.ListEntityAssetIDsByCreator(isLatestFirst, pageSize, bookmark)

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, resourceIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *IdentityController) handleGetAuthPendingList(ctx *gin.Context) {
	// Extract and check parameters
	pageSizeStr := ctx.Query("pageSize")
	bookmark := processBase64FromURLQuery(ctx.Query("bookmark"))

	pel := &ParameterErrorList{}
	pageSize := 10
	if strings.TrimSpace(pageSizeStr) != "" {
		pageSize = pel.AppendIfNotPositiveInt(pageSizeStr, "分页大小应为正整数。")
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// ListPendingAuthSessionIDsByResourceCreator
	authSessionIDs, err := c.AuthSvc.ListPendingAuthSessionIDsByResourceCreator(pageSize, bookmark)

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, authSessionIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *IdentityController) handleGetAuthRequestList(ctx *gin.Context) {
	// Extract and check parameters
	pageSizeStr := ctx.Query("pageSize")
	bookmark := processBase64FromURLQuery(ctx.Query("bookmark"))
	isLatestFirstStr := ctx.Query("isLatestFirst")

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
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// ListAuthSessionIDsByRequestor
	authSessionIDs, err := c.AuthSvc.ListAuthSessionIDsByRequestor(pageSize, bookmark, isLatestFirst)

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, authSessionIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}
