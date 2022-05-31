package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/idutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
	"github.com/XiaoYao-austin/ppks"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
)

// An EntityAssetController contains a group name and an `EntityService` instance. It also implements the interface `Controller`.
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
		urlMethodPair{"asset", "POST"}:             []gin.HandlerFunc{c.handleCreateAsset},
		urlMethodPair{"assets/list", "GET"}:        []gin.HandlerFunc{c.handleListAssetIDs},
		urlMethodPair{"asset/:id/metadata", "GET"}: []gin.HandlerFunc{c.handleGetAssetMetadata},
		urlMethodPair{"asset/:id", "GET"}:          []gin.HandlerFunc{c.handleGetAsset},
	}
}

func (c *EntityAssetController) handleCreateAsset(ctx *gin.Context) {
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
		} else {
			// The resource type can't be "Offchain"
			if resourceType == data.Offchain {
				*pel = append(*pel, "资源类型不能为链下。")
			}
		}
	}

	// Extract and check common parameters
	name := ctx.PostForm("name")
	name = pel.AppendIfEmptyOrBlankSpaces(name, "实体资产名称不能为空。")

	designDocumentID := ctx.PostForm("designDocumentId")
	designDocumentID = pel.AppendIfEmptyOrBlankSpaces(designDocumentID, "设计文档 ID 不能为空。")

	// Check componentsIDs
	componentIDsBytes := []byte(ctx.PostForm("componentIds"))
	var componentIDs []string
	if len(componentIDsBytes) == 0 {
		*pel = append(*pel, "组件的序列号不能为空。")
	} else {
		err = json.Unmarshal(componentIDsBytes, &componentIDs)
		if err != nil {
			*pel = append(*pel, "组件的序列号不合法。")
		}
	}

	// Check policy if it's not a plain resource
	policy := ctx.PostForm("policy")
	if resourceType != data.Plain {
		if len(policy) == 0 {
			*pel = append(*pel, "策略不能为空。")
		}
	}

	// Whether the properties are public should be specified if it's not a plain asset
	isNamePublic := true
	if resourceType != data.Plain {
		isNamePublicStr := ctx.PostForm("isNamePublic")
		isNamePublic = pel.AppendIfNotBool(isNamePublicStr, "必须指定资产名称公开性。")
	}

	isDesignDocumentIDPublic := true
	if resourceType != data.Plain {
		isDesignDocumentIDPublicStr := ctx.PostForm("isDesignDocumentIdPublic")
		isDesignDocumentIDPublic = pel.AppendIfNotBool(isDesignDocumentIDPublicStr, "必须指定设计文档 ID 公开性。")
	}

	// Early return after extracting common parameters if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Generate an ID
	id, err := idutils.GenerateSnowflakeId()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// A symmetric key should be generated to encrypt the resource if the resourse type is Encrypted (later used in the service function and returned as part of the result).
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
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("无法序列化对称公钥。"))
			return
		}
	}

	// Wrap the collected info into an asset
	asset := &common.EntityAsset{
		ID:                       id,
		Name:                     name,
		DesignDocumentID:         designDocumentID,
		IsNamePublic:             isNamePublic,
		IsDesignDocumentIDPublic: isDesignDocumentIDPublic,
		ComponentIDs:             componentIDs,
	}

	// Invoke the service function according to the resource type
	var txID string
	switch resourceType {
	case data.Plain:
		txID, err = c.EntityAssetSvc.CreateEntityAsset(asset)
	case data.Encrypted:
		txID, err = c.EntityAssetSvc.CreateEncryptedEntityAsset(asset, key, policy)
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

func (c *EntityAssetController) handleGetAssetMetadata(ctx *gin.Context) {
	id := ctx.Param("id")

	// Validity check
	pel := &ParameterErrorList{}
	id = pel.AppendIfEmptyOrBlankSpaces(id, "实体 ID 不能为空。")

	// Early return if there's parameter error
	if len(*pel) != 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, *pel)
		return
	}

	resDataMetadata, err := c.EntityAssetSvc.GetEntityAssetMetadata(id)
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

func (c *EntityAssetController) handleGetAsset(ctx *gin.Context) {
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
		} else {
			// The resource type can't be "Offchain"
			if resourceType == data.Offchain {
				*pel = append(*pel, "资源类型不能为链下。")
			}
		}
	}

	// Extract and check entity asset ID
	id := ctx.Param("id")
	id = pel.AppendIfEmptyOrBlankSpaces(id, "实体资产 ID 不能为空。")

	// Extract conditional parameters
	var keySwitchSessionID string
	var numSharesExpected int

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Invoke the service function to get the metadata
	resDataMetadata, err := c.EntityAssetSvc.GetEntityAssetMetadata(id)
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
	var entityAsset *common.EntityAsset
	switch resourceType {
	case data.Plain:
		entityAsset, err = c.EntityAssetSvc.GetEntityAsset(id, resDataMetadata)
	case data.Encrypted:
		// Try to get the entity asset from the database first
		entityAsset, err = c.EntityAssetSvc.GetDecryptedEntityAssetFromDB(id, resDataMetadata)
		if errors.Cause(err) == errorcode.ErrorNotFound || reflect.TypeOf(err) == reflect.TypeOf(&service.ErrorCorruptedDatabaseResult{}) {
			// Perform the full process if the document is not available in the database (not found or corrupted)
			// First try to get additional parameters
			keySwitchSessionID = ctx.Query("keySwitchSessionId")
			keySwitchSessionID = pel.AppendIfEmptyOrBlankSpaces(keySwitchSessionID, "该实体资产解密记录不可用，密钥置换会话 ID 不能为空。")

			numSharesExpectedString := ctx.Query("numSharesExpected")
			numSharesExpected = pel.AppendIfNotInt(numSharesExpectedString, "该实体资产解密记录不可用，期待的份额数量应为正整数。")

			// Early return if the error list is not empty
			if len(*pel) > 0 {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
				return
			}

			// Invoke the service function to perform the full process
			entityAsset, err = c.EntityAssetSvc.GetEncryptedEntityAsset(id, keySwitchSessionID, numSharesExpected, resDataMetadata)
		}
	}

	// Check error type and generate the corresponding response
	if err == nil {
		ctx.JSON(http.StatusOK, entityAsset)
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

func (c *EntityAssetController) handleListAssetIDs(ctx *gin.Context) {
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
	if temp := strings.TrimSpace(ctx.Query("resourceId")); temp != "" {
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

	var designDocumentID *string
	if temp := strings.TrimSpace(ctx.Query("designDocumentId")); temp != "" {
		designDocumentID = &temp
	}

	// Early return if the error list is not empty
	if len(*pel) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, pel)
		return
	}

	// Encapsulate the query conditions into a struct
	queryConditions := common.EntityAssetQueryConditions{
		CommonQueryConditions: common.CommonQueryConditions{
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
		DesignDocumentID: designDocumentID,
	}

	// Perform the query using the service function
	resourceIDs, err := c.EntityAssetSvc.ListEntityAssetIDsByConditions(&queryConditions, pageSize)

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
