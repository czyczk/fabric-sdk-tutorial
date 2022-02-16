package controller

import (
	"net/http"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// A DocumentController contains a group name and a `DocumentService` instance. It also implements the interface `Controller`.
type AuthController struct {
	GroupName string
	AuthSvc   service.AuthServiceInterface
}

// GetGroupName returns the group name.
func (c *AuthController) GetGroupName() string {
	return c.GroupName
}

// GetEndpointMap implements part of the interface `Controller`. It returns the API endpoints and handlers which are defined and managed by DocumentController.
func (c *AuthController) GetEndpointMap() EndpointMap {
	return EndpointMap{
		urlMethodPair{"request", "POST"}:  []gin.HandlerFunc{c.handleCreateAuthRequest},
		urlMethodPair{"response", "POST"}: []gin.HandlerFunc{c.handleCreateAuthResponse},
		urlMethodPair{":id", "GET"}:       []gin.HandlerFunc{c.handleGetAuthSession},
	}
}

func (c *AuthController) handleCreateAuthRequest(ctx *gin.Context) {
	// Validity check
	pel := &ParameterErrorList{}

	resourceID := ctx.PostForm("resourceId")
	resourceID = pel.AppendIfEmptyOrBlankSpaces(resourceID, "资源 ID 不能为空。")

	reason := ctx.PostForm("reason")

	// Early return after extracting common parameters if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	txID, err := c.AuthSvc.CreateAuthRequest(resourceID, reason)

	// Check error type and generate the corresponding response
	// The symmetric key will be included if it's not empty
	if err == nil {
		info := TransactionIDInfo{
			TransactionID: txID,
		}
		ctx.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *AuthController) handleCreateAuthResponse(ctx *gin.Context) {
	// Validity check
	pel := &ParameterErrorList{}

	authSessionID := ctx.PostForm("authSessionId")
	authSessionID = pel.AppendIfEmptyOrBlankSpaces(authSessionID, "授权会话 ID 不能为空。")

	result := ctx.PostForm("result")
	result = pel.AppendIfEmptyOrBlankSpaces(result, "授权结果不能为空。")
	resultBool := pel.AppendIfNotBool(result, "授权结果须为 bool 值。")

	// Early return after extracting common parameters if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	txID, err := c.AuthSvc.CreateAuthResponse(authSessionID, resultBool)
	// Check error type and generate the corresponding response
	// The symmetric key will be included if it's not empty
	if err == nil {
		info := TransactionIDInfo{
			TransactionID: txID,
		}
		ctx.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else if errors.Cause(err) == errorcode.ErrorForbidden {
		ctx.Writer.WriteHeader(http.StatusForbidden)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *AuthController) handleGetAuthSession(ctx *gin.Context) {
	// Validity check
	pel := &ParameterErrorList{}
	id := pel.AppendIfEmptyOrBlankSpaces(ctx.Param("id"), "授权会话 ID 不能为空")

	// Early return if there's parameter error
	if len(*pel) != 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, *pel)
		return
	}

	authSession, err := c.AuthSvc.GetAuthSession(id)
	if err == nil {
		ctx.JSON(http.StatusOK, authSession)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		ctx.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}
