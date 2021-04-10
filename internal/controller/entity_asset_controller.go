package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
	"github.com/XiaoYao-austin/ppks"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
)

// A EntityAssetController contains a group name and a `EntityService` instance. It also implements the interface `Controller`.
type EntityAssetController struct {
	GroupName      string
	EntityAssetSvc service.EntityAssetServiceInterface
}

// GetGroupName returns the group name.
func (c *EntityAssetController) GetGroupName() string {
	return c.GroupName
}

// GetEndpointMap implements part of the interface `Controller`. It returns the API endpoints and handlers which are defined and managed by EntityAssetController.
func (c *EntityAssetController) GetEndpointMap() EndpointMap {
	return EndpointMap{
		urlMethodPair{"", "POST"}:             []gin.HandlerFunc{c.handleCreateAsset},
		urlMethodPair{":id/metadata", "GET"}:  []gin.HandlerFunc{c.handleGetAssetMetadata},
		urlMethodPair{":id", "GET"}:           []gin.HandlerFunc{c.handleGetAsset},
		urlMethodPair{":id/transfer", "POST"}: []gin.HandlerFunc{c.handleTransferAsset},
	}
}

func (ec *EntityAssetController) handleCreateAsset(c *gin.Context) {
	resourceTypeStr := c.PostForm("resourceType")

	// Validity check
	pel := &ParameterErrorList{}

	resourceTypeStr = pel.AppendIfEmptyOrBlankSpaces(resourceTypeStr, "资源类型不能为空。")

	resourceType, err := data.NewResourceTypeFromString(resourceTypeStr)
	if err != nil {
		*pel = append(*pel, "资源类型不合法。")
	}
	if resourceType == data.Offchain || resourceType == data.RegulatorEncrypted {
		*pel = append(*pel, "资源类型不能为链下和监管者加密文档")
	}

	// Extract and check common parameters
	name := c.PostForm("name")
	name = pel.AppendIfEmptyOrBlankSpaces(name, "文档名称不能为空。")

	// Property is optional, but it must be valid (can be unmarshaled to a map) if provided.
	propertyBytes := []byte(c.PostForm("property"))
	property := make(map[string]string)
	if len(propertyBytes) != 0 {
		err = json.Unmarshal(propertyBytes, &property)
		if err != nil {
			*pel = append(*pel, "属性字段不合法。")
		}
	}

	// Check componentsIDs
	componentsIDsByte := []byte(c.PostForm("componentsIDs"))
	if len(componentsIDsByte) == 0 {
		*pel = append(*pel, "组件的序列号不能为空。")
	}
	var componentsIDs []string
	err = json.Unmarshal(componentsIDsByte, &componentsIDs)
	if err != nil {
		*pel = append(*pel, "组件的序列号不合法。")
	}

	// Check policy if it's not a plain resource
	policy := c.PostForm("policy")

	if resourceType != data.Plain {
		if len(policy) == 0 {
			*pel = append(*pel, "策略不能为空。")
		}
	}

	// Early return after extracting common parameters if the error list is not empty
	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Generate an ID
	sfNode, err := snowflake.NewNode(1)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("无法生成 ID。"))
		return
	}
	id := sfNode.Generate().String()

	// A symmetric key should be generated to encrypt the document if the resourse type is Encrypted (later used in the service function and returned as part of the result).
	var key *ppks.CurvePoint
	// The key is now a `*ppks.CurvePoint`. Cast it to a `*sm2.PublicKey` so that it can be converted to PEM bytes
	var keyAsPublicKey *sm2.PublicKey
	var keyPEM []byte
	// Generate the key and convert it into the useful type
	if resourceType == data.Encrypted {
		key = ppks.GenPoint()
		keyAsPublicKey = (*sm2.PublicKey)(key)
		keyPEM, err = sm2keyutils.ConvertPublicKeyToPEM(keyAsPublicKey)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("无法序列化对称公钥。"))
			return
		}
	}

	// Invoke the service function according to the resource type
	var txID string
	switch resourceType {
	case data.Plain:
		txID, err = ec.EntityAssetSvc.CreateEntityAsset(id, name, componentsIDs, string(propertyBytes))
	case data.Encrypted:
		txID, err = ec.EntityAssetSvc.CreateEncryptedEntityAsset(id, name, componentsIDs, string(propertyBytes), key, policy)
	}

	// Check error type and generate the corresponding response
	// The symmetric key will be included if it's not empty
	if err == nil {
		info := ResourceCreationInfo{
			ResourceID:           id,
			TransactionID:        txID,
			SymmetricKeyMaterial: string(keyPEM),
		}
		c.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (ec *EntityAssetController) handleGetAssetMetadata(c *gin.Context) {
	id := c.Param("entityAssetID")

	// Validity check
	pel := &ParameterErrorList{}
	id = pel.AppendIfEmptyOrBlankSpaces(id, "实体 ID 不能为空。")

	// Early return if there's parameter error
	if len(*pel) != 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, *pel)
		return
	}

	resDataMetadata, err := ec.EntityAssetSvc.GetEntityAssetMetadata(id)
	if err == nil {
		c.JSON(http.StatusOK, resDataMetadata)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		c.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (ec *EntityAssetController) handleGetAsset(c *gin.Context) {
	resourceTypeStr := c.Query("resourceType")

	// Validity check
	pel := &ParameterErrorList{}

	resourceTypeStr = pel.AppendIfEmptyOrBlankSpaces(resourceTypeStr, "资源类型不能为空。")
	resourceType, err := data.NewResourceTypeFromString(resourceTypeStr)
	if err != nil {
		*pel = append(*pel, "资源类型不合法。")
	}
	if resourceType == data.Offchain || resourceType == data.RegulatorEncrypted {
		*pel = append(*pel, "资源类型不能为链下和监管者加密文档。")
	}

	// Extract and check entity asset ID
	id := c.Param("entityAssetID")
	id = pel.AppendIfEmptyOrBlankSpaces(id, "文档 ID 不能为空。")

	// Extract conditional parameters
	var keySwitchSessionID string
	var numSharesExpected int

	if resourceType == data.Encrypted {
		keySwitchSessionID = c.Query("keySwitchSessionID")
		keySwitchSessionID = pel.AppendIfEmptyOrBlankSpaces(keySwitchSessionID, "密钥置换会话 ID 不能为空。")

		numSharesExpectedString := c.Query("numSharesExpected")
		numSharesExpected = pel.AppendIfNotInt(numSharesExpectedString, "期待的份额数量应为正整数。")
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Invoke the service function according to the resource type
	var entityAsset *common.EntityAsset
	switch resourceType {
	case data.Plain:
		entityAsset, err = ec.EntityAssetSvc.GetEntityAsset(id)
	case data.Encrypted:
		entityAsset, err = ec.EntityAssetSvc.GetEncryptedEntityAsset(id, keySwitchSessionID, numSharesExpected)
	}

	// Check error type and generate the corresponding response
	if err == nil {
		c.JSON(http.StatusOK, entityAsset)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		c.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (ec *EntityAssetController) handleTransferAsset(c *gin.Context) {
	id := c.PostForm("entityAssetID")

	// check entity asset ID
	pel := &ParameterErrorList{}
	id = pel.AppendIfEmptyOrBlankSpaces(id, "实体 ID 不能为空。")

	// check new owner
	newOwner := c.PostForm("newOwner")
	newOwner = pel.AppendIfEmptyOrBlankSpaces(newOwner, "新的拥有者不能为空。")

	// Early return if there's parameter error
	if len(*pel) != 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, *pel)
		return
	}

	// Generate a transferRecordID
	sfNode, err := snowflake.NewNode(1)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("无法生成 ID。"))
		return
	}
	transferRecordID := sfNode.Generate().String()

	// Invoke the service function
	txID, err := ec.EntityAssetSvc.TransferEntityAsset(transferRecordID, id, newOwner)

	// Check error type and generate the corresponding response
	if err == nil {
		info := ResourceCreationInfo{
			ResourceID:    transferRecordID,
			TransactionID: txID,
		}
		c.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		c.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
