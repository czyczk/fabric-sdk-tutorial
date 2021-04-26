package controller

import (
	"encoding/base64"
	"net/http"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// A DocumentController contains a group name and a `DocumentService` instance. It also implements the interface `Controller`.
type KeySwitchController struct {
	GroupName    string
	KeySwitchSvc service.KeySwitchServiceInterface
}

// GetGroupName returns the group name.
func (kc *KeySwitchController) GetGroupName() string {
	return kc.GroupName
}

// GetEndpointMap implements part of the interface `Controller`. It returns the API endpoints and handlers which are defined and managed by DocumentController.
func (kc *KeySwitchController) GetEndpointMap() EndpointMap {
	return EndpointMap{
		urlMethodPair{"trigger", "POST"}:               []gin.HandlerFunc{kc.handleCreateKeySwitchTrigger},
		urlMethodPair{":id/results/list-await", "GET"}: []gin.HandlerFunc{kc.handleAwaitListKeySwitchResults},
	}
}

func (kc *KeySwitchController) handleCreateKeySwitchTrigger(c *gin.Context) {
	resourceID := c.PostForm("resourceID")

	// Validity check
	pel := &ParameterErrorList{}

	resourceID = pel.AppendIfEmptyOrBlankSpaces(resourceID, "资源 ID 不能为空。")

	// Extract and check common parameters
	authSessionID := c.PostForm("authSessionID")

	txID, err := kc.KeySwitchSvc.CreateKeySwitchTrigger(resourceID, authSessionID)

	// Check error type and generate the corresponding response
	// The symmetric key will be included if it's not empty
	if err == nil {
		info := TransactionIDInfo{
			TransactionID: txID,
		}
		c.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorForbidden {
		c.Writer.WriteHeader(http.StatusForbidden)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (kc *KeySwitchController) handleAwaitListKeySwitchResults(c *gin.Context) {
	// Extract and check parameters
	keySwitchSessionID := c.Param("id")

	// Validity check
	pel := &ParameterErrorList{}

	keySwitchSessionID = pel.AppendIfEmptyOrBlankSpaces(keySwitchSessionID, "密钥置换会话 ID 不能为空。")

	numExpected := c.Query("numExpected")
	numExpectedInt := pel.AppendIfNotPositiveInt(numExpected, "预期的份额数量必须为正整数。")

	timeout := c.Query("timeout")
	timeoutInt := 0
	if timeout != "" {
		timeoutInt = pel.AppendIfNotPositiveInt(timeout, "超时时限必须为正整数。")
	}

	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	var shares [][]byte
	var err error
	if timeout == "" {
		shares, err = kc.KeySwitchSvc.AwaitKeySwitchResults(keySwitchSessionID, numExpectedInt)
	} else {
		shares, err = kc.KeySwitchSvc.AwaitKeySwitchResults(keySwitchSessionID, numExpectedInt, timeoutInt)
	}

	// Check error type and generate the corresponding response
	if err == nil {
		var ret []string
		for _, shareBytes := range shares {
			shareAsBase64 := base64.StdEncoding.EncodeToString(shareBytes)
			ret = append(ret, shareAsBase64)
		}

		c.JSON(http.StatusOK, ret)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else if errors.Cause(err) == errorcode.ErrorGatewayTimeout {
		c.Writer.WriteHeader(http.StatusGatewayTimeout)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
