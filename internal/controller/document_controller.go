package controller

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

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
		urlMethodPair{"document", "POST"}:               []gin.HandlerFunc{c.handleCreateDocument},
		urlMethodPair{"documents/list", "GET"}:          []gin.HandlerFunc{c.handleListDocumentIDs},
		urlMethodPair{"document/:id/metadata", "GET"}:   []gin.HandlerFunc{c.handleGetDocumentMetadata},
		urlMethodPair{"document/:id/properties", "GET"}: []gin.HandlerFunc{c.handleGetDocumentProperties},
		urlMethodPair{"document/:id", "GET"}:            []gin.HandlerFunc{c.handleGetDocument},
	}
}

func (c *DocumentController) handleCreateDocument(ctx *gin.Context) {
	resourceTypeStr := ctx.PostForm("resourceType")

	// Validity check
	pel := &ParameterErrorList{}

	resourceTypeStr = pel.AppendIfEmptyOrBlankSpaces(resourceTypeStr, "资源类型不能为空。")
	var resourceType data.ResourceType
	var err error
	if resourceTypeStr != "" {
		resourceType, err = data.NewResourceTypeFromString(resourceTypeStr)
		if err != nil {
			*pel = append(*pel, "资源类型不合法。")
		}
	}

	// Extract and check common parameters
	name := ctx.PostForm("name")
	name = pel.AppendIfEmptyOrBlankSpaces(name, "文档名称不能为空。")

	documentTypeStr := strings.TrimSpace(ctx.PostForm("documentType"))
	documentTypeStr = pel.AppendIfEmptyOrBlankSpaces(documentTypeStr, "文档类型不能为空。")
	var documentType common.DocumentType
	if documentTypeStr != "" {
		documentType, err = common.NewDocumentTypeFromString(documentTypeStr)
		if err != nil {
			*pel = append(*pel, "文档类型不合法。")
		}
	}

	precedingDocumentID := ctx.PostForm("precedingDocumentID")
	headDocumentID := ctx.PostForm("headDocumentID")
	entityAssetID := ctx.PostForm("entityAssetID")

	// Whether the properties are public should be specified if it's not a plain resource
	isNamePublic := true
	if resourceType != data.Plain {
		isNamePublicStr := ctx.PostForm("isNamePublic")
		isNamePublic = pel.AppendIfNotBool(isNamePublicStr, "必须指定文档名称公开性。")
	}

	isDocumentTypePublic := true
	if resourceType != data.Plain {
		isDocumentTypePublicStr := ctx.PostForm("isDocumentTypePublic")
		isDocumentTypePublic = pel.AppendIfNotBool(isDocumentTypePublicStr, "必须指定文档类型公开性。")
	}

	isPrecedingDocumentIDPublic := true
	if resourceType != data.Plain {
		isPrecedingDocumentIDPublicStr := ctx.PostForm("isPrecedingDocumentIDPublic")
		isPrecedingDocumentIDPublic = pel.AppendIfNotBool(isPrecedingDocumentIDPublicStr, "必须指定前序文档 ID 公开性。")
	}

	isHeadDocumentIDPublic := true
	if resourceType != data.Plain {
		isHeadDocumentIDPublicStr := ctx.PostForm("isHeadDocumentIDPublic")
		isHeadDocumentIDPublic = pel.AppendIfNotBool(isHeadDocumentIDPublicStr, "必须指定头文档 ID 公开性。")
	}

	isEntityAssetIDPublic := true
	if resourceType != data.Plain {
		isEntityAssetIDPublicStr := ctx.PostForm("isEntityAssetIDPublic")
		isEntityAssetIDPublic = pel.AppendIfNotBool(isEntityAssetIDPublicStr, "必须指定相关实体资产 ID 公开性。")
	}

	// Check contents
	file, fileHeader, err := ctx.Request.FormFile("contents")
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if fileHeader.Size == 0 {
		*pel = append(*pel, "文档内容不能为空。")
	}

	// Check policy if it's not a plain resource
	policy := ctx.PostForm("policy")

	if resourceType != data.Plain {
		if len(policy) == 0 {
			*pel = append(*pel, "策略不能为空。")
		}
	}

	// Early return after extracting common parameters if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Read the contents from the multipart file
	contents, err := io.ReadAll(file)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err = file.Close(); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	file = nil
	runtime.GC()

	// Generate an ID
	sfNode, err := snowflake.NewNode(1)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("无法生成 ID。"))
		return
	}
	id := sfNode.Generate().String()

	// A symmetric key should be generated to encrypt the resource if the resourse type is Encrypted or Offchain (later used in the service function and returned as part of the result).
	var key *ppks.CurvePoint
	// The key is now a `*ppks.CurvePoint`. Cast it to a `*sm2.PublicKey` so that it can be converted to PEM bytes
	var keyAsPublicKey *sm2.PublicKey
	var keyPEM []byte
	// Generate the key and convert it into the useful type
	if resourceType == data.Encrypted || resourceType == data.Offchain {
		key = ppks.GenPoint()
		keyAsPublicKey = (*sm2.PublicKey)(key)
		keyPEM, err = sm2keyutils.ConvertPublicKeyToPEM(keyAsPublicKey)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("无法序列化对称公钥。"))
			return
		}
	}

	// Wrap the collected info into a document
	document := &common.Document{
		DocumentProperties: common.DocumentProperties{
			ID:                          id,
			Name:                        name,
			Type:                        documentType,
			PrecedingDocumentID:         precedingDocumentID,
			HeadDocumentID:              headDocumentID,
			EntityAssetID:               entityAssetID,
			IsNamePublic:                isNamePublic,
			IsTypePublic:                isDocumentTypePublic,
			IsPrecedingDocumentIDPublic: isPrecedingDocumentIDPublic,
			IsHeadDocumentIDPublic:      isHeadDocumentIDPublic,
			IsEntityAssetIDPublic:       isEntityAssetIDPublic,
		},
		Contents: contents,
	}

	// Invoke the service function according to the resource type
	var txID string
	switch resourceType {
	case data.Plain:
		txID, err = c.DocumentSvc.CreateDocument(document)
	case data.Encrypted:
		txID, err = c.DocumentSvc.CreateEncryptedDocument(document, key, policy)
	case data.Offchain:
		txID, err = c.DocumentSvc.CreateOffchainDocument(document, key, policy)
	}

	// Check error type and generate the corresponding response
	// The symmetric key will be included if it's not empty
	if err == nil {
		info := ResourceCreationInfo{
			ResourceID:           id,
			TransactionID:        txID,
			SymmetricKeyMaterial: string(keyPEM),
		}
		ctx.JSON(http.StatusOK, info)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *DocumentController) handleGetDocumentMetadata(ctx *gin.Context) {
	id := ctx.Param("id")

	// Validity check
	pel := &ParameterErrorList{}
	id = pel.AppendIfEmptyOrBlankSpaces(id, "数字文档 ID 不能为空。")

	// Early return if there's parameter error
	if len(*pel) != 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, *pel)
		return
	}

	resDataMetadata, err := c.DocumentSvc.GetDocumentMetadata(id)
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

func (c *DocumentController) handleGetDocumentProperties(ctx *gin.Context) {
	// Validity check
	pel := &ParameterErrorList{}

	// Extract and check document ID
	id := ctx.Param("id")
	id = pel.AppendIfEmptyOrBlankSpaces(id, "文档 ID 不能为空。")

	// Extract conditional parameters
	var keySwitchSessionID string
	var numSharesExpected int

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Invoke the service function to get the metadata
	resDataMetadata, err := c.DocumentSvc.GetDocumentMetadata(id)
	if err != nil {
		if errors.Cause(err) == errorcode.ErrorNotFound {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
			ctx.AbortWithStatus(http.StatusNotImplemented)
			return
		} else {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Invoke the service function to get the document properties from the database first
	documentProperties, err := c.DocumentSvc.GetDecryptedDocumentPropertiesFromDB(id, resDataMetadata)

	// Try 	documentProperties, err = c.DocumentSvc.GetDecryptedDocumentPropertiesFromDB(id, resDataMetadata)
	if errors.Cause(err) == errorcode.ErrorNotFound || reflect.TypeOf(err) == reflect.TypeOf(&service.ErrorCorruptedDatabaseResult{}) {
		// Perform the full process if the document is not available in the database (not found or corrupted)
		// First try to get additional parameters
		keySwitchSessionID = ctx.Query("keySwitchSessionID")
		keySwitchSessionID = pel.AppendIfEmptyOrBlankSpaces(keySwitchSessionID, "该数字文档解密记录不可用，密钥置换会话 ID 不能为空。")

		numSharesExpectedString := ctx.Query("numSharesExpected")
		numSharesExpected = pel.AppendIfNotInt(numSharesExpectedString, "该数字文档解密记录不可用，期待的份额数量应为正整数。")

		// Early return if the error list is not empty
		if len(*pel) > 0 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
			return
		}

		// Invoke the service function to perform the full process
		documentProperties, err = c.DocumentSvc.GetEncryptedDocumentProperties(id, keySwitchSessionID, numSharesExpected, resDataMetadata)
	}

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, documentProperties)
	} else if reflect.TypeOf(err) == reflect.TypeOf(&service.ErrorBadRequest{}) {
		*pel = append(*pel, err.Error())
		ctx.JSON(http.StatusBadRequest, pel)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		ctx.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *DocumentController) handleGetDocument(ctx *gin.Context) {
	resourceTypeStr := ctx.Query("resourceType")

	// Validity check
	pel := &ParameterErrorList{}

	resourceTypeStr = pel.AppendIfEmptyOrBlankSpaces(resourceTypeStr, "资源类型不能为空。")
	var resourceType data.ResourceType
	var err error
	if resourceTypeStr != "" {
		resourceType, err = data.NewResourceTypeFromString(resourceTypeStr)
		if err != nil {
			*pel = append(*pel, "资源类型不合法。")
		}
	}

	// Extract and check document ID
	id := ctx.Param("id")
	id = pel.AppendIfEmptyOrBlankSpaces(id, "文档 ID 不能为空。")

	// Extract conditional parameters
	var keySwitchSessionID string
	var numSharesExpected int

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Invoke the service function to get the metadata
	resDataMetadata, err := c.DocumentSvc.GetDocumentMetadata(id)
	if err != nil {
		if errors.Cause(err) == errorcode.ErrorNotFound {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
			ctx.AbortWithStatus(http.StatusNotImplemented)
			return
		} else {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Invoke the service function according to the resource type
	var document *common.Document
	switch resourceType {
	case data.Plain:
		document, err = c.DocumentSvc.GetDocument(id, resDataMetadata)
	default:
		// The same process for both "Encrypted" and "Offchain" documents
		// Try to get the document from the database first
		document, err = c.DocumentSvc.GetDecryptedDocumentFromDB(id, resDataMetadata)
		if errors.Cause(err) == errorcode.ErrorNotFound || reflect.TypeOf(err) == reflect.TypeOf(&service.ErrorCorruptedDatabaseResult{}) {
			// Perform the full process if the document is not available in the database (not found or corrupted)
			// First try to get additional parameters
			keySwitchSessionID = ctx.Query("keySwitchSessionID")
			keySwitchSessionID = pel.AppendIfEmptyOrBlankSpaces(keySwitchSessionID, "该数字文档解密记录不可用，密钥置换会话 ID 不能为空。")

			numSharesExpectedString := ctx.Query("numSharesExpected")
			numSharesExpected = pel.AppendIfNotInt(numSharesExpectedString, "该数字文档解密记录不可用，期待的份额数量应为正整数。")

			// Early return if the error list is not empty
			if len(*pel) > 0 {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
				return
			}

			// Invoke the service function to perform the full process
			if resourceType == data.Encrypted {
				document, err = c.DocumentSvc.GetEncryptedDocument(id, keySwitchSessionID, numSharesExpected, resDataMetadata)
			} else {
				document, err = c.DocumentSvc.GetOffchainDocument(id, keySwitchSessionID, numSharesExpected, resDataMetadata)
			}
		}
	}

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, document)
	} else if reflect.TypeOf(err) == reflect.TypeOf(&service.ErrorBadRequest{}) {
		*pel = append(*pel, err.Error())
		ctx.JSON(http.StatusBadRequest, pel)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		ctx.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

func (c *DocumentController) handleListDocumentIDs(ctx *gin.Context) {
	// Extract and check parameters
	pel := &ParameterErrorList{}
	var err error

	// Theses fields have their default values if not specified
	pageSizeStr := ctx.Query("pageSize")
	pageSize := 10
	if strings.TrimSpace(pageSizeStr) != "" {
		pageSize = pel.AppendIfNotPositiveInt(pageSizeStr, "分页大小应为正整数。")
	}

	bookmarkStr := strings.TrimSpace(ctx.Query("bookmark"))
	var bookmark *string
	if bookmarkStr != "" {
		bookmark = &bookmarkStr
	}

	isLatestFirst := true
	isLatestFirstStr := ctx.Query("isLatestFirst")
	if isLatestFirstStr != "" {
		isLatestFirst = pel.AppendIfNotBool(isLatestFirstStr, "最新于最前选项必须为 bool 值。")
	}

	// Optional fields
	var resourceID *string
	if temp := strings.TrimSpace(ctx.Query("resourceID")); temp != "" {
		resourceID = &temp
	}

	var isNameExact *bool
	if temp := strings.TrimSpace(ctx.Query("isNameExact")); temp != "" {
		tempBool := pel.AppendIfNotBool(temp, "是否为精确名称选项必须为 bool 值。")
		isNameExact = &tempBool
	}
	var name *string
	if temp := strings.TrimSpace(ctx.Query("name")); temp != "" {
		name = &temp
	}

	var isTimeExact *bool
	if temp := strings.TrimSpace(ctx.Query("isTimeExact")); temp != "" {
		tempBool := pel.AppendIfNotBool(temp, "是否为精确时间选项必须为 bool 值。")
		isTimeExact = &tempBool
	}
	var exactTime *time.Time
	if temp := strings.TrimSpace(ctx.Query("time")); temp != "" {
		tempTime := pel.AppendIfNotTime(temp, "时间应为合法的 RFC3339 格式。")
		exactTime = &tempTime
	}
	var timeAfterInclusive *time.Time
	if temp := strings.TrimSpace(ctx.Query("timeAfterInclusive")); temp != "" {
		tempTime := pel.AppendIfNotTime(temp, "开始时间应为合法的 RFC3339 格式。")
		timeAfterInclusive = &tempTime
	}
	var timeBeforeExclusive *time.Time
	if temp := strings.TrimSpace(ctx.Query("timeBeforeExclusive")); temp != "" {
		tempTime := pel.AppendIfNotTime(temp, "结束时间应为合法的 RFC3339 格式。")
		timeBeforeExclusive = &tempTime
	}

	var documentType *common.DocumentType
	if temp := strings.TrimSpace(ctx.Query("documentType")); temp != "" {
		tempDocumentType, err := common.NewDocumentTypeFromString(temp)
		if err != nil {
			*pel = append(*pel, "文档类型不合法。")
		} else {
			documentType = &tempDocumentType
		}
	}

	var precedingDocumentID *string
	if temp := strings.TrimSpace(ctx.Query("precedingDocumentID")); temp != "" {
		precedingDocumentID = &temp
	}

	var headDocumentID *string
	if temp := strings.TrimSpace(ctx.Query("headDocumentID")); temp != "" {
		headDocumentID = &temp
	}

	var entityAssetID *string
	if temp := strings.TrimSpace(ctx.Query("entityAssetID")); temp != "" {
		entityAssetID = &temp
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Encapsulate the query conditions into a struct
	queryConditions := service.DocumentQueryConditions{
		CommonQueryConditions: service.CommonQueryConditions{
			IsDesc:              isLatestFirst,
			ResourceID:          resourceID,
			IsNameExact:         isNameExact,
			Name:                name,
			IsTimeExact:         isTimeExact,
			Time:                exactTime,
			TimeAfterInclusive:  timeAfterInclusive,
			TimeBeforeExclusive: timeBeforeExclusive,
			LastResourceID:      bookmark,
		},
		DocumentType:        documentType,
		PrecedingDocumentID: precedingDocumentID,
		HeadDocumentID:      headDocumentID,
		EntityAssetID:       entityAssetID,
	}

	// Perform the query using the service function
	resourceIDs, err := c.DocumentSvc.ListDocumentIDsByConditions(queryConditions, pageSize)

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, resourceIDs)
	} else if errors.Cause(err) == errorcode.ErrorNotFound {
		ctx.Writer.WriteHeader(http.StatusNotFound)
	} else if errors.Cause(err) == errorcode.ErrorNotImplemented {
		ctx.Writer.WriteHeader(http.StatusNotImplemented)
	} else {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}
