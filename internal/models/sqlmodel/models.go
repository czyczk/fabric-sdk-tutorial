package sqlmodel

import (
	"database/sql"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gorm.io/gorm"
)

// Document 定义了数据库表 documents，用于读写数据库中的数字文档。
type Document struct {
	gorm.Model
	ID                          string
	Name                        string
	PrecedingDocumentID         sql.NullString
	HeadDocumentID              sql.NullString
	EntityAssetID               sql.NullString
	IsNamePublic                bool
	IsPrecedingDocumentIDPublic bool
	IsHeadDocumentIDPublic      bool
	IsEntityAssetIDPublic       bool
	Contents                    []byte
}

// EntityAsset 定义了数据库表 entity_assets，用于读写数据库中的实体资产。
type EntityAsset struct {
	gorm.Model
	ID                       string
	Name                     string
	DesignDocumentID         string
	IsNamePublic             bool
	IsDesignDocumentIDPublic bool
	Components               []Component
}

// Component 定义了 EntityAsset 的组成部分，用于读写数据库中的实体资产。EntityAsset 与 Component 为一对多关系。
type Component struct {
	gorm.Model
	ID            string
	EntityAssetID string
}

// ToModel 将一个 `sqlmodel.Document` 对象转为 `common.Document` 对象。
func (d *Document) ToModel() common.Document {
	ret := common.Document{
		ID:                          d.ID,
		Name:                        d.Name,
		PrecedingDocumentID:         d.PrecedingDocumentID.String,
		HeadDocumentID:              d.HeadDocumentID.String,
		EntityAssetID:               d.EntityAssetID.String,
		IsNamePublic:                d.IsNamePublic,
		IsPrecedingDocumentIDPublic: d.IsPrecedingDocumentIDPublic,
		IsHeadDocumentIDPublic:      d.IsHeadDocumentIDPublic,
		IsEntityAssetIDPublic:       d.IsEntityAssetIDPublic,
		Contents:                    d.Contents,
	}

	return ret
}

// ToModel 将一个 `sqlmodel.EntityAsset` 对象转为 `common.EntityAsset` 对象。
func (e *EntityAsset) ToModel() common.EntityAsset {
	componentIDs := make([]string, len(e.Components))
	for i, component := range e.Components {
		componentIDs[i] = component.ID
	}

	ret := common.EntityAsset{
		ID:                       e.ID,
		Name:                     e.Name,
		DesignDocumentID:         e.DesignDocumentID,
		IsNamePublic:             e.IsNamePublic,
		IsDesignDocumentIDPublic: e.IsDesignDocumentIDPublic,
		ComponentIDs:             componentIDs,
	}

	return ret
}

// NewDocumentFromModel 通过 `common.Document` 对象创建一个 `sqlmodel.Document` 对象。
func NewDocumentFromModel(model common.Document) Document {
	var precedingDocumentID sql.NullString
	_ = precedingDocumentID.Scan(model.PrecedingDocumentID)
	var headDocumentID sql.NullString
	_ = headDocumentID.Scan(model.HeadDocumentID)
	var entityAssetID sql.NullString
	_ = entityAssetID.Scan(model.EntityAssetID)

	ret := Document{
		ID:                          model.ID,
		Name:                        model.Name,
		PrecedingDocumentID:         precedingDocumentID,
		HeadDocumentID:              headDocumentID,
		EntityAssetID:               entityAssetID,
		IsNamePublic:                model.IsNamePublic,
		IsPrecedingDocumentIDPublic: model.IsPrecedingDocumentIDPublic,
		IsHeadDocumentIDPublic:      model.IsHeadDocumentIDPublic,
		Contents:                    model.Contents,
	}

	return ret
}

// NewEntityAssetFromModel 通过 `common.EntityAsset` 对象创建一个 `sqlmodel.EntityAsset` 对象。
func NewEntityAssetFromModel(model common.EntityAsset) EntityAsset {
	components := make([]Component, len(model.ComponentIDs))
	for i, componentID := range model.ComponentIDs {
		component := Component{
			ID:            componentID,
			EntityAssetID: model.ID,
		}
		components[i] = component
	}

	ret := EntityAsset{
		ID:                       model.ID,
		Name:                     model.Name,
		DesignDocumentID:         model.DesignDocumentID,
		IsNamePublic:             model.IsNamePublic,
		IsDesignDocumentIDPublic: model.IsDesignDocumentIDPublic,
		Components:               components,
	}

	return ret
}
