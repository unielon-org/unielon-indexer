package storage

import (
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
)

func (c *DBClient) CrossDeploy(tx *gorm.DB, cross *models.CrossInfo) error {

	//WDOGE(WRAPPED-DOGE)
	tick := "W" + cross.Tick + "(WRAPPED-" + cross.Tick + ")"
	drc20c := &models.Drc20Collect{
		Tick:          tick,
		Max:           models.NewNumber(-1),
		Dec:           8,
		HolderAddress: cross.HolderAddress,
		TxHash:        cross.TxHash,
	}

	err := tx.Create(drc20c).Error
	if err != nil {
		return err
	}

	cc := &models.CrossCollect{
		Tick:          cross.Tick,
		AdminAddress:  cross.AdminAddress,
		HolderAddress: cross.HolderAddress,
	}

	err = tx.Create(cc).Error
	if err != nil {
		return err
	}

	revert := &models.CrossRevert{
		Op:          "deploy",
		Tick:        cross.Tick,
		BlockNumber: cross.BlockNumber,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) CrossMint(tx *gorm.DB, cross *models.CrossInfo) error {

	err := c.MintDrc20(tx, cross.Tick, cross.ToAddress, cross.Amt.Int(), cross.TxHash, cross.BlockNumber, false)
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) CrossBurn(tx *gorm.DB, cross *models.CrossInfo) error {
	err := c.BurnDrc20(tx, cross.Tick, cross.HolderAddress, cross.Amt.Int(), cross.TxHash, cross.BlockNumber, false)
	if err != nil {
		return err
	}
	return nil
}
