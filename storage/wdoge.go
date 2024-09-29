package storage

import (
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
)

func (c *DBClient) DogeDeposit(tx *gorm.DB, wdoge *models.WDogeInfo) error {

	err := c.MintDrc20(tx, wdoge.Tick, wdoge.HolderAddress, wdoge.Amt.Int(), wdoge.BlockNumber, false)
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) DogeWithdraw(tx *gorm.DB, wdoge *models.WDogeInfo) error {

	err := c.BurnDrc20(tx, wdoge.Tick, wdoge.HolderAddress, wdoge.Amt.Int(), wdoge.BlockNumber, false)
	if err != nil {
		return err
	}

	return nil
}
