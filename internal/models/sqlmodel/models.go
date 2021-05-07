package sqlmodel

import (
	"database/sql"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Document 定义了数据库表 documents，用于读写数据库中的数字文档。
type Document struct {
	gorm.Model
	ID                          int64
	Name                        string
	Type                        common.DocumentType
	PrecedingDocumentID         sql.NullInt64
	HeadDocumentID              sql.NullInt64
	EntityAssetID               sql.NullInt64
	IsNamePublic                bool
	IsTypePublic                bool
	IsPrecedingDocumentIDPublic bool
	IsHeadDocumentIDPublic      bool
	IsEntityAssetIDPublic       bool
	Contents                    []byte
}

// EntityAsset 定义了数据库表 entity_assets，用于读写数据库中的实体资产。
type EntityAsset struct {
	gorm.Model
	ID                       int64
	Name                     string
	DesignDocumentID         int64
	IsNamePublic             bool
	IsDesignDocumentIDPublic bool
	Components               []Component
}

// Component 定义了 EntityAsset 的组成部分，用于读写数据库中的实体资产。EntityAsset 与 Component 为一对多关系。
type Component struct {
	gorm.Model
	ID            int64
	EntityAssetID int64
}

// ToModel 将一个 `sqlmodel.Document` 对象转为 `common.Document` 对象。
func (d *Document) ToModel() *common.Document {
	ret := &common.Document{
		ID:                          parseInt64ToSnowflakeString(d.ID),
		Name:                        d.Name,
		PrecedingDocumentID:         parseNullInt64ToSnowflakeString(d.PrecedingDocumentID),
		HeadDocumentID:              parseNullInt64ToSnowflakeString(d.HeadDocumentID),
		EntityAssetID:               parseNullInt64ToSnowflakeString(d.EntityAssetID),
		IsNamePublic:                d.IsNamePublic,
		IsPrecedingDocumentIDPublic: d.IsPrecedingDocumentIDPublic,
		IsHeadDocumentIDPublic:      d.IsHeadDocumentIDPublic,
		IsEntityAssetIDPublic:       d.IsEntityAssetIDPublic,
		Contents:                    d.Contents,
	}

	return ret
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

// NewDocumentFromModel 通过 `common.Document` 对象创建一个 `sqlmodel.Document` 对象。
func NewDocumentFromModel(model *common.Document) (*Document, error) {
	errMsg := "无法转换数字文档对象为数据库对象"

	precedingDocumentID, err := parseSnowflakeStringToNullInt64(model.PrecedingDocumentID)
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	headDocumentID, err := parseSnowflakeStringToNullInt64(model.HeadDocumentID)
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	entityAssetID, err := parseSnowflakeStringToNullInt64(model.EntityAssetID)
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	id, err := parseSnowflakeStringToInt64(model.ID)
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	ret := &Document{
		ID:                          id,
		Name:                        model.Name,
		PrecedingDocumentID:         precedingDocumentID,
		HeadDocumentID:              headDocumentID,
		EntityAssetID:               entityAssetID,
		IsNamePublic:                model.IsNamePublic,
		IsPrecedingDocumentIDPublic: model.IsPrecedingDocumentIDPublic,
		IsHeadDocumentIDPublic:      model.IsHeadDocumentIDPublic,
		Contents:                    model.Contents,
	}

	return ret, nil
}

// NewEntityAssetFromModel 通过 `common.EntityAsset` 对象创建一个 `sqlmodel.EntityAsset` 对象。
func NewEntityAssetFromModel(model *common.EntityAsset) (*EntityAsset, error) {
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
		IsNamePublic:             model.IsNamePublic,
		IsDesignDocumentIDPublic: model.IsDesignDocumentIDPublic,
		Components:               components,
	}

	return ret, nil
}
