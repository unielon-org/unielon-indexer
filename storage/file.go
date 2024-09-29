package storage

import (
	"fmt"
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
)

func (c *DBClient) FileDeploy(tx *gorm.DB, model *models.FileInfo) error {

	fileCollectAddress := &models.FileCollectAddress{
		FileId:        model.FileId,
		FilePath:      model.FilePath,
		HolderAddress: model.HolderAddress,
	}

	err := tx.Create(fileCollectAddress).Error
	if err != nil {
		return fmt.Errorf("deploy InstallNftCollect err: %s order_id: %s", err.Error(), model.OrderId)
	}

	revert := &models.FileRevert{
		FromAddress: "",
		ToAddress:   model.ToAddress,
		FileId:      model.FileId,
		BlockNumber: model.BlockNumber,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return fmt.Errorf("deploy InstallNftRevert err: %s order_id: %s", err.Error(), model.OrderId)
	}

	return nil
}

func (c *DBClient) FileTransfer(tx *gorm.DB, model *models.FileInfo) error {
	err := c.TransferFile(tx, model.HolderAddress, model.ToAddress, model.FileId, model.BlockNumber, false)
	if err != nil {
		return fmt.Errorf("transfer err: %s order_id: %s", err, model.OrderId)
	}

	return nil
}
