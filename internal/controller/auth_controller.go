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
	}
}

func (ac *AuthController) handleCreateAuthRequest(c *gin.Context) {
	// Validity check
	pel := &ParameterErrorList{}

	resourceID := c.PostForm("resourceID")
	resourceID = pel.AppendIfEmptyOrBlankSpaces(resourceID, "资源 ID 不能为空。")

	reason := c.PostForm("reason")

	// Early return after extracting common parameters if the error list is not empty
	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	txID, err := ac.AuthSvc.CreateAuthRequest(resourceID, reason)

	// Check error type and generate the corresponding response
	// The symmetric key will be included if it's not empty
	if err == nil {
		info := TransactionIDInfo{
			TransactionID: txID,
		}
		c.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (ac *AuthController) handleCreateAuthResponse(c *gin.Context) {
	// Validity check
	pel := &ParameterErrorList{}

	authSessionID := c.PostForm("authSessionID")
	authSessionID = pel.AppendIfEmptyOrBlankSpaces(authSessionID, "授权会话 ID 不能为空。")

	result := c.PostForm("result")
	result = pel.AppendIfEmptyOrBlankSpaces(result, "授权结果不能为空。")
	resultBool := pel.AppendIfNotBool(result, "授权结果须为 bool 值。")

	// Early return after extracting common parameters if the error list is not empty
	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	txID, err := ac.AuthSvc.CreateAuthResponse(authSessionID, resultBool)
	// Check error type and generate the corresponding response
	// The symmetric key will be included if it's not empty
	if err == nil {
		info := TransactionIDInfo{
			TransactionID: txID,
		}
		c.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else if errors.Cause(err) == errorcode.ErrorForbidden {
		c.Writer.WriteHeader(http.StatusForbidden)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
