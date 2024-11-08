package storage

import (
	"fmt"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/gorm"
	"math/big"
)

func (c *DBClient) BoxDeploy(tx *gorm.DB, box *models.BoxInfo, reservesAddress string) error {

	drc20c := &models.Drc20Collect{
		Tick:          box.Tick0,
		Max:           box.Max,
		Dec:           8,
		HolderAddress: box.HolderAddress,
		TxHash:        box.TxHash,
	}

	err := tx.Create(drc20c).Error
	if err != nil {
		return fmt.Errorf("BoxDeploy err: %s order_id: %s", err.Error(), box.OrderId)
	}

	if err := c.MintDrc20(tx, box.Tick0, reservesAddress, box.Max.Int(), box.TxHash, box.BlockNumber, false); err != nil {
		return fmt.Errorf("BoxDeploy MintDrc20 err: %s", err.Error())
	}

	boxCollect := &models.BoxCollect{
		Tick0:           box.Tick0,
		Tick1:           box.Tick1,
		Max:             box.Max,
		Amt0:            box.Amt0,
		LiqAmt:          box.LiqAmt,
		LiqBlock:        box.LiqBlock,
		Amt1:            box.Amt1,
		HolderAddress:   box.HolderAddress,
		ReservesAddress: reservesAddress,
	}

	err = tx.Create(boxCollect).Error
	if err != nil {
		return fmt.Errorf("BoxCollect Create err: %s order_id: %s", err.Error(), box.OrderId)
	}

	revert := &models.BoxRevert{
		Op:          "deploy",
		Tick0:       box.Tick0,
		BlockNumber: box.BlockNumber,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return fmt.Errorf("BoxRevert Create err: %s order_id: %s", err.Error(), box.OrderId)
	}

	return nil
}

func (c *DBClient) BoxMint(tx *gorm.DB, box *models.BoxInfo, reservesAddress string) error {

	boxc := &models.BoxCollect{}
	err := tx.Where("tick0 = ?", box.Tick0).First(boxc).Error
	if err != nil {
		return fmt.Errorf("BoxMint FindBoxCollectByTick err: %s", err.Error())
	}

	err = c.TransferDrc20(tx, boxc.Tick1, box.HolderAddress, reservesAddress, box.Amt1.Int(), box.TxHash, box.BlockNumber, false)
	if err != nil {
		return err
	}

	err = tx.Exec(`UPDATE box_collect
				SET liqamt_finish = (
					SELECT b.amt_sum
					FROM drc20_collect_address b
					WHERE 
						box_collect.tick1 = b.tick AND 
						box_collect.reserves_address = b.holder_address
				)
				WHERE EXISTS (
					SELECT 1
					FROM drc20_collect_address b
					WHERE 
						box_collect.is_del = 0 AND 
						box_collect.reserves_address = ? AND 
						box_collect.tick1 = ?
				);`, reservesAddress, boxc.Tick1).Error

	if err != nil {
		return err
	}

	ba := &models.BoxCollectAddress{
		Tick:          boxc.Tick0,
		Amt:           box.Amt1,
		HolderAddress: box.HolderAddress,
		BlockNumber:   box.BlockNumber,
	}

	err = tx.Create(ba).Error
	if err != nil {
		return fmt.Errorf("BoxMint Create err: %s", err.Error())
	}

	boxc1 := &models.BoxCollect{}
	err = tx.Where("tick0 = ?", box.Tick0).First(boxc1).Error
	if err != nil {
		return fmt.Errorf("BoxMint FindBoxCollectByTick err: %s", err.Error())
	}

	if boxc1.LiqAmt.Int().Cmp(big.NewInt(0)) > 0 && boxc1.LiqAmtFinish.Cmp(boxc1.LiqAmt) >= 0 {
		err := c.BoxFinish(tx, boxc1, box.BlockNumber)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBClient) BoxFinish(tx *gorm.DB, boxc *models.BoxCollect, height int64) error {
	swap := &models.SwapInfo{
		Op:            "create",
		Tick0:         boxc.Tick0,
		Tick1:         boxc.Tick1,
		Amt0:          boxc.Amt0,
		Amt1:          boxc.LiqAmtFinish,
		HolderAddress: boxc.ReservesAddress,
		TxHash:        "box_finish",
		BlockNumber:   height,
	}

	swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, _, _ = utils.SortTokens(swap.Tick0, swap.Tick1, swap.Amt0, swap.Amt1, nil, nil)

	swap.Amt0Out = swap.Amt0
	swap.Amt1Out = swap.Amt1

	err := c.SwapCreate(tx, swap)
	if err != nil {
		return fmt.Errorf("swapCreate SwapCreate error: %v", err)
	}

	total := big.NewInt(0)
	bas := []*models.BoxCollectAddress{}
	err = tx.Where("tick = ?", boxc.Tick0).Find(&bas).Error
	if err != nil {
		return err
	}

	for _, ba := range bas {
		total = total.Add(total, ba.Amt.Int())
	}

	for _, ba := range bas {
		amt := big.NewInt(0).Div(big.NewInt(0).Mul(ba.Amt.Int(), boxc.Amt0.Int()), total)
		err = c.TransferDrc20(tx, boxc.Tick0, boxc.ReservesAddress, ba.HolderAddress, amt, "box_finish", height, false)
		if err != nil {
			return err
		}
	}

	err = tx.Model(&models.BoxCollect{}).Where("tick0 = ?", boxc.Tick0).Update("amt0_finish", boxc.Amt0).Error
	if err != nil {
		return err
	}

	revert := &models.BoxRevert{
		Op:          "finish",
		Tick0:       swap.Tick0,
		Tick1:       swap.Tick1,
		BlockNumber: height,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return err
	}

	return nil
}

func (c *DBClient) BoxRefund(tx *gorm.DB, boxc *models.BoxCollect, height int64) error {

	err := c.BurnDrc20(tx, boxc.Tick0, boxc.ReservesAddress, boxc.Max.Int(), "box_refund", height, false)
	if err != nil {
		return err
	}

	drc20c := &models.Drc20Collect{}
	err = tx.Model(&models.Drc20Collect{}).Where("tick = ?", boxc.Tick0).First(drc20c).Error
	if err != nil {
		return err
	}

	revert := &models.BoxRevert{
		Op:            "refund-drc20",
		Tick0:         boxc.Tick0,
		Max:           drc20c.Max,
		HolderAddress: boxc.ReservesAddress,
		TxHash:        "box_refund",
		BlockNumber:   height,
	}

	err = tx.Create(revert).Error
	if err != nil {
		return err
	}

	err = tx.Delete(&models.Drc20Collect{}, "tick = ?", boxc.Tick0).Error
	if err != nil {
		return err
	}

	err = tx.Model(&models.BoxCollect{}).Where("tick0 = ?", boxc.Tick0).Update("is_del", 1).Error
	if err != nil {
		return err
	}

	bas := []*models.BoxCollectAddress{}
	err = tx.Where("tick = ?", boxc.Tick0).Find(&bas).Error
	if err != nil {
		return err
	}

	for _, ba := range bas {
		err = c.TransferDrc20(tx, boxc.Tick1, boxc.ReservesAddress, ba.HolderAddress, ba.Amt.Int(), "box_refund", height, false)
		if err != nil {
			return err
		}
	}

	return nil

}
