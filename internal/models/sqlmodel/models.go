package sqlmodel

import (
	"database/sql"
	"fmt"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Document 定义了数据库表 documents 的属性部分，用于读写数据库中的数字文档的属性部分。
type DocumentProperties struct {
	gorm.Model
	ID                          int64
	Name                        string `gorm:"type:VARCHAR(255) NOT NULL"`
	Type                        string `gorm:"type:ENUM('DESIGN_DOCUMENT', 'PRODUCTION_DOCUMENT', 'TRANSFER_DOCUMENT', 'USAGE_DOCUMENT', 'REPAIR_DOCUMENT') NOT NULL"`
	PrecedingDocumentID         sql.NullInt64
	HeadDocumentID              sql.NullInt64
	EntityAssetID               sql.NullInt64
	TimeCreated                 time.Time `gorm:"not null"`
	IsNamePublic                bool      `gorm:"not null"`
	IsTypePublic                bool      `gorm:"not null"`
	IsPrecedingDocumentIDPublic bool      `gorm:"not null"`
	IsHeadDocumentIDPublic      bool      `gorm:"not null"`
	IsEntityAssetIDPublic       bool      `gorm:"not null"`
}

// Document 定义了数据库表 documents，用于读写数据库中的数字文档。
type Document struct {
	gorm.Model
	ID       int64
	Contents []byte
}

// EntityAsset 定义了数据库表 entity_assets，用于读写数据库中的实体资产。
type EntityAsset struct {
	gorm.Model
	ID                       int64
	Name                     string      `gorm:"type:VARCHAR(255) NOT NULL"`
	DesignDocumentID         int64       `gorm:"not null"`
	TimeCreated              time.Time   `gorm:"not null"`
	IsNamePublic             bool        `gorm:"not null"`
	IsDesignDocumentIDPublic bool        `gorm:"not null"`
	Components               []Component `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Component 定义了 EntityAsset 的组成部分，用于读写数据库中的实体资产。EntityAsset 与 Component 为一对多关系。
type Component struct {
	gorm.Model
	ID            int64
	EntityAssetID int64 `gorm:"not null"`
}

// 自定义 DocumentProperties 的表名。
func (DocumentProperties) TableName() string {
	return "document_properties"
}

// ToModel 将一个 `sqlmodel.DocumentProperties` 对象转为 `common.DocumentProperties` 对象。
func (d *DocumentProperties) ToModel() *common.DocumentProperties {
	ret := &common.DocumentProperties{
		ID:                          parseInt64ToSnowflakeString(d.ID),
		Name:                        d.Name,
		Type:                        getDocumentTypeFromSQLValue(d.Type),
		PrecedingDocumentID:         parseNullInt64ToSnowflakeString(d.PrecedingDocumentID),
		HeadDocumentID:              parseNullInt64ToSnowflakeString(d.HeadDocumentID),
		EntityAssetID:               parseNullInt64ToSnowflakeString(d.EntityAssetID),
		IsNamePublic:                d.IsNamePublic,
		IsTypePublic:                d.IsTypePublic,
		IsPrecedingDocumentIDPublic: d.IsPrecedingDocumentIDPublic,
		IsHeadDocumentIDPublic:      d.IsHeadDocumentIDPublic,
		IsEntityAssetIDPublic:       d.IsEntityAssetIDPublic,
	}

	return ret
}

// ToModel 将一个 `sqlmodel.Document` 对象转为 `common.Document` 对象。
func (d *Document) ToModel(p *DocumentProperties) (*common.Document, error) {
	if d.ID != p.ID {
		return nil, fmt.Errorf("文档与文档属性的 ID 不匹配")
	}

	ret := &common.Document{
		DocumentProperties: *p.ToModel(),
		Contents:           d.Contents,
	}

	return ret, nil
}

// ToModel 将一个 `sqlmodel.EntityAsset` 对象转为 `common.EntityAsset` 对象。
func (e *EntityAsset) ToModel() *common.EntityAsset {
	componentIDs := make([]string, len(e.Components))
	for i, component := range e.Components {
		componentIDs[i] = parseInt64ToSnowflakeString(component.ID)
	}

	ret := &common.EntityAsset{
		ID:                       parseInt64ToSnowflakeString(e.ID),
		Name:                     e.Name,
		DesignDocumentID:         parseInt64ToSnowflakeString(e.DesignDocumentID),
		IsNamePublic:             e.IsNamePublic,
		IsDesignDocumentIDPublic: e.IsDesignDocumentIDPublic,
		ComponentIDs:             componentIDs,
	}

	return ret
}

// NewDocumentPropertiesFromModel 通过 `common.DocumentProperties` 对象创建一个 `sqlmodel.DocumentProperties` 对象。
func NewDocumentPropertiesFromModel(model *common.DocumentProperties, timeCreated time.Time) (*DocumentProperties, error) {
	errMsg := "无法转换数字文档属性对象为数据库对象"

	precedingDocumentID, err := parseSnowflakeStringToNullInt64(model.PrecedingDocumentID)
	if err != nil {
		return nil, errors.Wrapf(err, errMsg+": precedingDocumentID: %v", precedingDocumentID)
	}

	headDocumentID, err := parseSnowflakeStringToNullInt64(model.HeadDocumentID)
	if err != nil {
		return nil, errors.Wrapf(err, errMsg+": headDocumentID: %v", headDocumentID)
	}

	entityAssetID, err := parseSnowflakeStringToNullInt64(model.EntityAssetID)
	if err != nil {
		return nil, errors.Wrapf(err, errMsg+": entityAssetID: %v", entityAssetID)
	}

	id, err := parseSnowflakeStringToInt64(model.ID)
	if err != nil {
		return nil, errors.Wrapf(err, errMsg+": id: %v", id)
	}

	ret := &DocumentProperties{
		ID:                          id,
		Name:                        model.Name,
		Type:                        getSQLValueFromDocumentType(model.Type),
		PrecedingDocumentID:         precedingDocumentID,
		HeadDocumentID:              headDocumentID,
		EntityAssetID:               entityAssetID,
		TimeCreated:                 timeCreated,
		IsNamePublic:                model.IsNamePublic,
		IsTypePublic:                model.IsTypePublic,
		IsPrecedingDocumentIDPublic: model.IsPrecedingDocumentIDPublic,
		IsHeadDocumentIDPublic:      model.IsHeadDocumentIDPublic,
		IsEntityAssetIDPublic:       model.IsEntityAssetIDPublic,
	}

	return ret, nil
}

// NewDocumentFromModel 通过 `common.Document` 对象创建一个 `sqlmodel.Document` 对象和一个 `sqlmodel.DocumentProperties` 对象。
func NewDocumentFromModel(model *common.Document, timeCreated time.Time) (document *Document, documentProperties *DocumentProperties, err error) {
	errMsg := "无法转换数字文档对象为数据库对象"

	documentProperties, err = NewDocumentPropertiesFromModel(&model.DocumentProperties, timeCreated)
	if err != nil {
		err = errors.Wrap(err, errMsg)
		return
	}

	id, err := parseSnowflakeStringToInt64(model.ID)
	if err != nil {
		err = errors.Wrap(err, errMsg)
		return
	}

	document = &Document{
		ID:       id,
		Contents: model.Contents,
	}

	return
}

// NewEntityAssetFromModel 通过 `common.EntityAsset` 对象创建一个 `sqlmodel.EntityAsset` 对象。
func NewEntityAssetFromModel(model *common.EntityAsset, timeCreated time.Time) (*EntityAsset, error) {
	errMsg := "无法转换实体资产对象为数据库对象"

	entityAssetID, err := parseSnowflakeStringToInt64(model.ID)
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	components := make([]Component, len(model.ComponentIDs))
	for i, componentID := range model.ComponentIDs {
		id, err := parseSnowflakeStringToInt64(componentID)
		if err != nil {
			return nil, errors.Wrap(err, errMsg)
		}

		component := Component{
			ID:            id,
			EntityAssetID: entityAssetID,
		}
		components[i] = component
	}

	designDocumentID, err := parseSnowflakeStringToInt64(model.DesignDocumentID)
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	ret := &EntityAsset{
		ID:                       entityAssetID,
		Name:                     model.Name,
		DesignDocumentID:         designDocumentID,
		TimeCreated:              timeCreated,
		IsNamePublic:             model.IsNamePublic,
		IsDesignDocumentIDPublic: model.IsDesignDocumentIDPublic,
		Components:               components,
	}

	return ret, nil
}

func getSQLValueFromDocumentType(t common.DocumentType) string {
	switch t {
	case common.DesignDocument:
		return "DESIGN_DOCUMENT"
	case common.ProductionDocument:
		return "PRODUCTION_DOCUMENT"
	case common.TransferDocument:
		return "TRANSFER_DOCUMENT"
	case common.UsageDocument:
		return "USAGE_DOCUMENT"
	case common.RepairDocument:
		return "REPAIR_DOCUMENT"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

func getDocumentTypeFromSQLValue(v string) common.DocumentType {
	switch v {
	case "DESIGN_DOCUMENT":
		return common.DesignDocument
	case "PRODUCTION_DOCUMENT":
		return common.ProductionDocument
	case "TRANSFER_DOCUMENT":
		return common.TransferDocument
	case "USAGE_DOCUMENT":
		return common.UsageDocument
	case "REPAIR_DOCUMENT":
		return common.RepairDocument
	default:
		panic("未识别的文档类型")
	}
}
