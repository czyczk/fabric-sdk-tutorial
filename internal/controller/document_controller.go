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

// A DocumentController contains a group name and a `DocumentService` instance. It also implements the interface `Controller`.
type DocumentController struct {
	GroupName   string
	DocumentSvc service.DocumentServiceInterface
}

// GetGroupName returns the group name.
func (c *DocumentController) GetGroupName() string {
	return c.GroupName
}

// GetEndpointMap implements part of the interface `Controller`. It returns the API endpoints and handlers which are defined and managed by DocumentController.
func (c *DocumentController) GetEndpointMap() EndpointMap {
	return EndpointMap{
		urlMethodPair{"", "POST"}:            []gin.HandlerFunc{c.handleCreateDocument},
		urlMethodPair{":id/metadata", "GET"}: []gin.HandlerFunc{c.handleGetDocumentMetadata},
		urlMethodPair{":id", "GET"}:          []gin.HandlerFunc{c.handleGetDocument},
	}
}

func (dc *DocumentController) handleCreateDocument(c *gin.Context) {
	resourceTypeStr := c.PostForm("resourceType")

	// Validity check
	pel := &ParameterErrorList{}

	resourceTypeStr = pel.AppendIfEmptyOrBlankSpaces(resourceTypeStr, "资源类型不能为空。")

	resourceType, err := data.NewResourceTypeFromString(resourceTypeStr)
	if err != nil {
		*pel = append(*pel, "资源类型不合法。")
	}

	// Extract and check common parameters
	name := c.PostForm("name")

	// Property is optional, but it must be valid (can be unmarshaled to a map) if provided.
	propertyBytes := []byte(c.PostForm("property"))
	property := make(map[string]string)
	if len(propertyBytes) != 0 {
		err = json.Unmarshal(propertyBytes, &property)
		if err != nil {
			*pel = append(*pel, "属性字段不合法。")
		}
	}

	// Check contents if it's not an offchain resource
	contents := []byte(c.PostForm("contents"))
	if resourceType != data.Offchain {
		if len(contents) == 0 {
			*pel = append(*pel, "文档内容不能为空。")
		}
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

	// A symmetric key should be generated to encrypt the document if the resourse type is Encrypted or Offchain (later used in the service function and returned as part of the result).
	var key *ppks.CurvePoint
	// The key is now a `*ppks.CurvePoint`. Cast it to a `*sm2.PublicKey` so that it can be converted to PEM bytes
	keyAsPublicKey := (*sm2.PublicKey)(key)
	var keyPEM []byte
	// Generate the key and convert it into the useful type
	if resourceType == data.Encrypted || resourceType == data.Offchain {
		key = ppks.GenPoint()

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
		txID, err = dc.DocumentSvc.CreateDocument(id, name, contents, string(propertyBytes))
	case data.Encrypted:
		txID, err = dc.DocumentSvc.CreateEncryptedDocument(id, name, contents, string(propertyBytes), key, policy)
	case data.RegulatorEncrypted:
		txID, err = dc.DocumentSvc.CreateRegulatorEncryptedDocument(id, name, contents, string(propertyBytes), key)
	case data.Offchain:
		txID, err = dc.DocumentSvc.CreateOffchainDocument(id, name, string(propertyBytes), key, policy)
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

func (dc *DocumentController) handleGetDocumentMetadata(c *gin.Context) {
	id := c.Param("id")

	// Validity check
	pel := &ParameterErrorList{}
	id = pel.AppendIfEmptyOrBlankSpaces(id, "数字文档 ID 不能为空。")

	// Early return if there's parameter error
	if len(*pel) != 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, *pel)
		return
	}

	resDataMetadata, err := dc.DocumentSvc.GetDocumentMetadata(id)
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

func (dc *DocumentController) handleGetDocument(c *gin.Context) {
	resourceTypeStr := c.Query("resourceType")

	// Validity check
	pel := &ParameterErrorList{}

	resourceTypeStr = pel.AppendIfEmptyOrBlankSpaces(resourceTypeStr, "资源类型不能为空。")
	resourceType, err := data.NewResourceTypeFromString(resourceTypeStr)
	if err != nil {
		*pel = append(*pel, "资源类型不合法。")
	}

	// Early return if the resource type is "Offchain"
	if resourceType == data.Offchain {
		*pel = append(*pel, "资源类型不能为链下。")
		c.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Extract and check document ID
	id := c.Param("id")
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
	var document *common.Document
	switch resourceType {
	case data.Plain:
		document, err = dc.DocumentSvc.GetDocument(id)
	case data.Encrypted:
		document, err = dc.DocumentSvc.GetEncryptedDocument(id, keySwitchSessionID, numSharesExpected)
	case data.RegulatorEncrypted:
		document, err = dc.DocumentSvc.GetRegulatorEncryptedDocument(id)
	}

	// Check error type and generate the corresponding response
	if err == nil {
		c.JSON(http.StatusOK, document)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		c.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		c.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
