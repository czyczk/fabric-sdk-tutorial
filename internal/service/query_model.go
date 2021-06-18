package service

import (
	"fmt"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gorm.io/gorm"
)

type QueryConditions interface {
	ToCouchDBConditions() (conditions map[string]interface{}, err error)
	ToGormConditionedDB(db *gorm.DB) (tx *gorm.DB)
}

// CommonQueryConditions 表示适用所有通用模型的查询条件。不单独使用，用于组合于其他查询条件。
type CommonQueryConditions struct {
	IsDesc              bool
	ResourceID          *string
	IsNameExact         *bool // 名称是否为精确名称。`nil` 表示条件不启用。`true` 为精确名称，`false` 为部分名称（名称关键字）。该字段若不为 `nil` 则 `Name` 字段不可为 `nil`。
	Name                *string
	IsTimeExact         *bool // 时间是否为精确时间。`nil` 表示条件不启用。`true` 为精确时间，`false` 为时间范围。该字段若不为 `nil` 则相应的时间字段（精确时间字段或范围时间字段）不可为 `nil`。
	Time                *time.Time
	TimeAfterInclusive  *time.Time // 时间条件启用时，`TimeAfterInclusive` 和 `TimeBeforeExclusive` 不可全为 `nil`。
	TimeBeforeExclusive *time.Time // 时间条件启用时，`TimeAfterInclusive` 和 `TimeBeforeExclusive` 不可全为 `nil`。
	LastResourceID      *string    // 上一次查询最后的资源 ID
}

// DocumentQueryConditions 表示适用数字文档的查询条件。
type DocumentQueryConditions struct {
	CommonQueryConditions
	DocumentType        *common.DocumentType
	PrecedingDocumentID *string
	HeadDocumentID      *string
	EntityAssetID       *string
}

// EntityAssetQueryConditions 表示适用实体资产的查询条件。
type EntityAssetQueryConditions struct {
	CommonQueryConditions
	DesignDocumentID *string
}

func (c *CommonQueryConditions) ToCouchDBConditions() (conditions map[string]interface{}, err error) {
	conditions = make(map[string]interface{})
	conditions["selector"] = make(map[string]interface{})

	resourceIDSort := "asc"
	if c.IsDesc {
		resourceIDSort = "desc"
	}

	conditions["sort"] = []interface{}{
		map[string]string{
			"resourceID": resourceIDSort,
		},
	}

	if c.ResourceID != nil {
		conditions["selector"].(map[string]interface{})["resourceID"] = *c.ResourceID
	} else if c.LastResourceID != nil {
		if !c.IsDesc {
			conditions["selector"].(map[string]interface{})["resourceID"] = map[string]interface{}{
				"$gt": *c.LastResourceID,
			}
		} else {
			conditions["selector"].(map[string]interface{})["resourceID"] = map[string]interface{}{
				"$lt": *c.LastResourceID,
			}
		}
	}

	if c.IsNameExact != nil {
		if c.Name == nil {
			err = fmt.Errorf("名称条件启用时，必须指定名称字段")
			return
		}

		if *c.IsNameExact {
			conditions["selector"].(map[string]interface{})["extensions.name"] = *c.Name
		} else {
			conditions["selector"].(map[string]interface{})["extensions.name"] = map[string]interface{}{
				"$regex": *c.Name,
			}
		}
	}

	if c.IsTimeExact != nil {
		if *c.IsTimeExact {
			if c.Time == nil {
				err = fmt.Errorf("精确时间条件启用时，必须指定精确时间字段")
				return
			}

			conditions["selector"].(map[string]interface{})["timestamp"] = *c.Time
		} else {
			if c.TimeAfterInclusive == nil && c.TimeBeforeExclusive == nil {
				err = fmt.Errorf("时间范围条件启用时，必须指定至少一个时间范围字段")
				return
			}

			conditions["selector"].(map[string]interface{})["timestamp"] = make(map[string]interface{})

			if c.TimeAfterInclusive != nil {
				conditions["selector"].(map[string]interface{})["timestamp"].(map[string]interface{})["$gt"] = *c.TimeAfterInclusive
			}
			if c.TimeBeforeExclusive != nil {
				conditions["selector"].(map[string]interface{})["timestamp"].(map[string]interface{})["$le"] = *c.TimeBeforeExclusive
			}
		}
	}

	return
}

func (c *CommonQueryConditions) ToGormConditionedDB(db *gorm.DB) (tx *gorm.DB, err error) {
	tx = db

	resourceIDSort := ""
	if c.IsDesc {
		resourceIDSort = " desc"
	}
	tx = tx.Order("id" + resourceIDSort)

	if c.ResourceID != nil {
		tx = tx.Where("id = ?", *c.ResourceID)
	} else if c.LastResourceID != nil {
		if !c.IsDesc {
			tx = tx.Where("id > ?", *c.LastResourceID)
		} else {
			tx = tx.Where("id < ?", *c.LastResourceID)
		}
	}

	if c.IsNameExact != nil {
		if c.Name == nil {
			err = fmt.Errorf("名称条件启用时，必须指定名称字段")
			return
		}

		if *c.IsNameExact {
			tx = tx.Where("name = ?", *c.Name)
		} else {
			tx = tx.Where("name LIKE ?", fmt.Sprintf("%%%v%%", *c.Name))
		}
	}

	if c.IsTimeExact != nil {
		if *c.IsTimeExact {
			if c.Time == nil {
				err = fmt.Errorf("精确时间条件启用时，必须指定精确时间字段")
				return
			}

			tx = tx.Where("time_created = ?", *c.Time)
		} else {
			if c.TimeAfterInclusive == nil && c.TimeBeforeExclusive == nil {
				err = fmt.Errorf("时间范围条件启动时，必须指定至少一个时间范围字段")
				return
			}

			if c.TimeAfterInclusive != nil {
				tx = tx.Where("time_created >= ?", *c.TimeAfterInclusive)
			}
			if c.TimeBeforeExclusive != nil {
				tx = tx.Where("time_created < ?", *c.TimeBeforeExclusive)
			}
		}
	}

	return
}

func (c *DocumentQueryConditions) ToCouchDBConditions() (conditions map[string]interface{}, err error) {
	conditions, err = c.CommonQueryConditions.ToCouchDBConditions()
	if err != nil {
		return
	}

	conditions["selector"].(map[string]interface{})["extensions.dataType"] = "document"

	if c.DocumentType != nil {
		conditions["selector"].(map[string]interface{})["extensions.documentType"] = c.DocumentType
	}

	if c.PrecedingDocumentID != nil {
		conditions["selector"].(map[string]interface{})["extensions.precedingDocumentID"] = *c.PrecedingDocumentID
	}

	if c.HeadDocumentID != nil {
		conditions["selector"].(map[string]interface{})["extensions.headDocumentID"] = *c.HeadDocumentID
	}

	if c.EntityAssetID != nil {
		conditions["selector"].(map[string]interface{})["extensions.entityAssetID"] = *c.EntityAssetID
	}

	return
}

func (c *DocumentQueryConditions) ToGormConditionedDB(db *gorm.DB) (tx *gorm.DB, err error) {
	tx, err = c.CommonQueryConditions.ToGormConditionedDB(db)
	if err != nil {
		return
	}

	if c.DocumentType != nil {
		tx = tx.Where("type = ?", *c.DocumentType)
	}

	if c.PrecedingDocumentID != nil {
		tx = tx.Where("preceding_document_id = ?", *c.PrecedingDocumentID)
	}

	if c.HeadDocumentID != nil {
		tx = tx.Where("head_document_id = ?", *c.HeadDocumentID)
	}

	if c.EntityAssetID != nil {
		tx = tx.Where("entity_asset_id = ?", *c.EntityAssetID)
	}

	return
}

func (c *EntityAssetQueryConditions) ToCouchDBConditions() (conditions map[string]interface{}, err error) {
	conditions, err = c.CommonQueryConditions.ToCouchDBConditions()
	if err != nil {
		return
	}

	conditions["selector"].(map[string]interface{})["extensions.dataType"] = "entityAsset"

	if c.DesignDocumentID != nil {
		conditions["selector"].(map[string]interface{})["extensions.designDocumentID"] = *c.DesignDocumentID
	}

	return
}

func (c *EntityAssetQueryConditions) ToGormConditionedDB(db *gorm.DB) (tx *gorm.DB, err error) {
	tx, err = c.CommonQueryConditions.ToGormConditionedDB(db)
	if err != nil {
		return
	}

	if c.DesignDocumentID != nil {
		tx = tx.Where("design_document_id = ?", *c.DesignDocumentID)
	}

	return
}
