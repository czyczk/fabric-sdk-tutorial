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

// SaveDecryptedEntityAssetToLocalDB 将 `common.EntityAsset` 保存到指定的数据库中。
func SaveDecryptedEntityAssetToLocalDB(entityAsset *common.EntityAsset, timeCreated time.Time, db *gorm.DB) error {
	entityAssetDB, err := sqlmodel.NewEntityAssetFromModel(entityAsset, timeCreated)
	if err != nil {
		return err
	}

	// 写入或覆盖实体资产于相应表
	// dbResult := s.ServiceInfo.DB.Where("`entity_asset_id` = ?", id).Delete(&sqlmodel.Component{})
	dbResult := db.Exec("DELETE FROM `components` WHERE `entity_asset_id` = ?", entityAsset.ID)
	if dbResult.Error != nil {
		return errors.Wrap(dbResult.Error, "无法将解密后的实体资产存入数据库")
	}

	dbResult = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(entityAssetDB)
	if dbResult.Error != nil {
		return errors.Wrap(dbResult.Error, "无法将解密后的实体资产存入数据库")
	}

	return nil
}

// GetDecryptedEntityAssetFromLocalDB 从数据库中读取指定 ID 的实体资产。
func GetDecryptedEntityAssetFromLocalDB(id string, db *gorm.DB) (*sqlmodel.EntityAsset, error) {
	// 从数据库中读取解密后的实体资产
	var assetDB sqlmodel.EntityAsset
	dbResult := db.Where("id = ?", id).Take(&assetDB)
	if dbResult.Error != nil {
		if errors.Cause(dbResult.Error) == gorm.ErrRecordNotFound {
			return nil, errorcode.ErrorNotFound
		} else {
			return nil, errors.Wrap(dbResult.Error, "无法从数据库中获取资产")
		}
	}

	return &assetDB, nil
}
