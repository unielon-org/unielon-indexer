package storage

import (
	"fmt"
	"github.com/unielon-org/unielon-indexer/models"
	"gorm.io/gorm"
)

func (e *DBClient) FileExchangeCreate(tx *gorm.DB, ex *models.FileExchangeInfo, reservesAddress string) error {

	exc := &models.FileExchangeCollect{
		ExId:            ex.ExId,
		FileId:          ex.FileId,
		Tick:            ex.Tick,
		Amt:             ex.Amt,
		AmtFinish:       new(models.Number),
		HolderAddress:   ex.HolderAddress,
		ReservesAddress: reservesAddress,
	}

	file := &models.FileCollectAddress{}
	err := tx.Where("file_id = ?", ex.FileId).First(file).Error
	if err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	nft := &models.NftCollectAddress{}
		//	err = tx.Where("deploy_hash = ?", ex.FileId).First(nft).Error
		//	if err != nil {
		//		return fmt.Errorf("the contract does not exist err %s", err.Error())
		//	}
		//
		//	exc.IsNft = 1
		//
		//	err = e.TransferNft(tx, nft.Tick, exc.HolderAddress, exc.ReservesAddress, nft.TickId, ex.BlockNumber, false)
		//	if err != nil {
		//		return err
		//	}
		//} else {
		return fmt.Errorf("the contract does not exist err %s", err.Error())
		//}
	}

	err = e.TransferFile(tx, exc.HolderAddress, exc.ReservesAddress, exc.FileId, ex.TxHash, ex.BlockNumber, false)
	if err != nil {
		return err
	}

	exr := &models.FileExchangeRevert{
		Op:          "create",
		ExId:        ex.ExId,
		FileId:      ex.FileId,
		Tick:        ex.Tick,
		Amt:         ex.Amt,
		BlockNumber: ex.BlockNumber,
		IsNft:       exc.IsNft,
	}

	err = tx.Create(exc).Error
	if err != nil {
		return err
	}

	err = tx.Create(exr).Error
	if err != nil {
		return err
	}

	return nil
}

func (e *DBClient) FileExchangeTrade(tx *gorm.DB, ex *models.FileExchangeInfo) error {

	var exc *models.FileExchangeCollect
	err := tx.Where("ex_id = ?", ex.ExId).First(&exc).Error
	if exc == nil {
		return fmt.Errorf("file_exchange_collect not found")
	}

	err = e.TransferDrc20(tx, exc.Tick, ex.HolderAddress, exc.HolderAddress, exc.Amt.Int(), ex.TxHash, ex.BlockNumber, false)
	if err != nil {
		return err
	}

	err = e.TransferFile(tx, exc.ReservesAddress, ex.HolderAddress, exc.FileId, ex.TxHash, ex.BlockNumber, false)
	if err != nil {
		return err
	}

	err = tx.Model(exc).Update("amt_finish", exc.Amt.String()).Error
	if err != nil {
		return err
	}

	exr := &models.FileExchangeRevert{
		Op:          "trade",
		ExId:        ex.ExId,
		FileId:      ex.FileId,
		Tick:        ex.Tick,
		Amt:         ex.Amt,
		AmtFinish:   exc.Amt,
		BlockNumber: ex.BlockNumber,
		IsNft:       exc.IsNft,
	}

	err = tx.Create(exr).Error
	if err != nil {
		return err
	}

	return nil

}

func (e *DBClient) FileExchangeCancel(tx *gorm.DB, ex *models.FileExchangeInfo) error {

	var exc *models.FileExchangeCollect
	err := tx.Where("ex_id = ?", ex.ExId).First(&exc).Error
	if exc == nil {
		return fmt.Errorf("file_exchange_collect not found")
	}

	err = e.TransferFile(tx, exc.ReservesAddress, exc.HolderAddress, exc.FileId, ex.TxHash, ex.BlockNumber, false)
	if err != nil {
		return err
	}

	err = tx.Model(exc).Update("amt_finish", exc.Amt.String()).Error
	if err != nil {
		return err
	}

	exr := &models.FileExchangeRevert{
		Op:          "cancel",
		ExId:        ex.ExId,
		FileId:      ex.FileId,
		Tick:        ex.Tick,
		Amt:         ex.Amt,
		AmtFinish:   exc.Amt,
		BlockNumber: ex.BlockNumber,
		IsNft:       exc.IsNft,
	}

	err = tx.Create(exr).Error
	if err != nil {
		return err
	}

	return nil
}
