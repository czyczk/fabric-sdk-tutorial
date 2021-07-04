package db

import (
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/sqlmodel"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveDecryptedDocumentAndDocumentPropertiesToLocalDB 将 `common.Document` 对象保存到指定的数据库中。
func SaveDecryptedDocumentAndDocumentPropertiesToLocalDB(document *common.Document, timeCreated time.Time, db *gorm.DB) error {
	// 开一个交易，将解密的文档属性和内容存入数据库（若已存在则覆盖）
	err := db.Transaction(func(tx *gorm.DB) error {
		documentDB, documentPropertiesDB, err := sqlmodel.NewDocumentFromModel(document, timeCreated)
		if err != nil {
			return err
		}

		// 写入或覆盖文档属性部分于 document_properties 表
		dbResult := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(documentPropertiesDB)
		if dbResult.Error != nil {
			return errors.Wrap(dbResult.Error, "无法将解密后的文档属性存入数据库")
		}

		// 写入或覆盖文档内容部分于 documents 表
		dbResult = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(documentDB)
		if dbResult.Error != nil {
			return errors.Wrap(dbResult.Error, "无法将解密后的文档内容存入数据库")
		}

		return nil
	})

	return err
}

// SaveDecryptedDocumentPropertiesToLocalDB 将 `common.DocumentProperties` 保存到指定的数据库中。
func SaveDecryptedDocumentPropertiesToLocalDB(documentProperties *common.DocumentProperties, timeCreated time.Time, db *gorm.DB) error {
	documentPropertiesDB, err := sqlmodel.NewDocumentPropertiesFromModel(documentProperties, timeCreated)
	if err != nil {
		return err
	}

	// 写入或覆盖文档属性部分于 document_properties 表
	dbResult := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(documentPropertiesDB)
	if dbResult.Error != nil {
		return errors.Wrap(dbResult.Error, "无法将解密后的文档属性存入数据库")
	}

	return nil
}

// GetDecryptedDocumentContentsFromLocalDB 从数据库中读取指定 ID 的数字文档的内容部分。
func GetDecryptedDocumentContentsFromLocalDB(id string, db *gorm.DB) (*sqlmodel.Document, error) {
	// 从数据库中读取解密后的文档内容部分
	var documentDB sqlmodel.Document
	dbResult := db.Where("id = ?", id).Take(&documentDB)
	if dbResult.Error != nil {
		if errors.Cause(dbResult.Error) == gorm.ErrRecordNotFound {
			return nil, errorcode.ErrorNotFound
		} else {
			return nil, errors.Wrap(dbResult.Error, "无法从数据库中获取文档内容")
		}
	}

	return &documentDB, nil
}

// GetDecryptedDocumentPropertiesFromLocalDB 从数据库中读取指定 ID 的数字文档的属性部分。
func GetDecryptedDocumentPropertiesFromLocalDB(id string, db *gorm.DB) (*sqlmodel.DocumentProperties, error) {
	// 从数据库中读取解密后的文档属性部分
	var documentPropertiesDB sqlmodel.DocumentProperties
	dbResult := db.Where("id = ?", id).Take(&documentPropertiesDB)
	if dbResult.Error != nil {
		if errors.Cause(dbResult.Error) == gorm.ErrRecordNotFound {
			return nil, errorcode.ErrorNotFound
		} else {
			return nil, errors.Wrap(dbResult.Error, "无法从数据库中获取文档属性")
		}
	}

	return &documentPropertiesDB, nil
}
