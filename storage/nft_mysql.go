package storage

import (
	"database/sql"
	"fmt"
	"github.com/unielon-org/unielon-indexer/utils"
)

func (c *DBClient) InstallNftInfo(nft *utils.NFTInfo) error {
	query := "INSERT INTO nft_info (order_id, op, tick, tick_id, total, model, prompt, image, fee_address, fee_address_all, holder_address, to_address, fee_tx_hash, nft_tx_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := c.SqlDB.Exec(query, nft.OrderId, nft.Op, nft.Tick, nft.TickId, nft.Total, nft.Model, nft.Prompt, nft.Image, nft.FeeAddress, nft.FeeAddressAll, nft.HolderAddress, nft.ToAddress, nft.FeeTxHash, nft.NftTxHash)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *DBClient) InstallNftRevert(tx *sql.Tx, tick, from, to string, tickId int64, height int64, prompt, image, deployHash string) error {
	exec := "INSERT INTO nft_revert (tick, from_address, to_address, tick_id, block_number, prompt, image, deploy_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, tick, from, to, tickId, height, prompt, image, deployHash)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateNftInfo(swap *utils.NFTInfo) error {
	query := "update nft_info set fee_tx_hash = ?,  fee_tx_index = ?, fee_block_hash = ?, fee_block_number = ?, nft_tx_hash = ?, nft_tx_raw = ? where order_id = ?"
	_, err := c.SqlDB.Exec(query, swap.FeeTxHash, swap.FeeTxIndex, swap.FeeBlockHash, swap.FeeBlockNumber, swap.NftTxHash, swap.NftTxRaw, swap.OrderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) UpdateNftInfoFork(tx *sql.Tx, height int64) error {
	query := "update nft_info set nft_block_number = 0, nfte_block_hash = '', order_status = 0 where nft_block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) InstallNftCollect(tx *sql.Tx, tick string, total int64, model, prompt, image, holder_address, deploy_hash string) error {
	query := "INSERT INTO nft_collect (tick, total, model, prompt, image, holder_address, deploy_hash) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err := tx.Exec(query, tick, total, model, prompt, image, holder_address, deploy_hash)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) FindNftInfoByFee(feeAddress string) (*utils.NFTInfo, error) {
	query := "SELECT order_id, op, tick, tick_id, total, model, prompt, image, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, nft_tx_hash, nft_block_hash, nft_block_number, fee_address, holder_address, to_address, err_info, order_status, update_date, create_date  FROM nft_info where fee_address = ? and fee_tx_hash = ''"
	rows, err := c.SqlDB.Query(query, feeAddress)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		nft := &utils.NFTInfo{}
		err := rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &nft.TickId, &nft.Total, &nft.Model, &nft.Prompt, &nft.Image, &nft.FeeTxHash, &nft.FeeTxIndex, &nft.FeeBlockHash, &nft.FeeBlockNumber, &nft.NftTxHash, &nft.NftBlockHash, &nft.NftBlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.ToAddress, &nft.ErrInfo, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)
		if err != nil {
			return nil, err
		}
		nft.ImageData, _ = utils.Base64ToPng(nft.Image)
		return nft, nil
	}
	return nil, nil
}

func (c *DBClient) FindNftInfoByTxHash(NftTXHash string) (*utils.NFTInfo, error) {
	query := "SELECT order_id, op, tick, tick_id, total, model, prompt, image, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, nft_tx_hash, nft_block_hash, nft_block_number, fee_address, holder_address,  to_address, err_info, order_status, update_date, create_date  FROM nft_info where nft_tx_hash = ?"
	rows, err := c.SqlDB.Query(query, NftTXHash)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		nft := &utils.NFTInfo{}
		err := rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &nft.TickId, &nft.Total, &nft.Model, &nft.Prompt, &nft.Image, &nft.FeeTxHash, &nft.FeeTxIndex, &nft.FeeBlockHash, &nft.FeeBlockNumber, &nft.NftTxHash, &nft.NftBlockHash, &nft.NftBlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.ToAddress, &nft.ErrInfo, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)
		if err != nil {
			return nil, err
		}
		return nft, nil
	}
	return nil, nil
}

func (c *DBClient) FindNftInfoById(OrderId string) (*utils.NFTInfo, error) {
	query := "SELECT  order_id, op, tick, tick_id, total, model, prompt, image, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, nft_tx_hash, nft_block_hash, nft_block_number, fee_address, holder_address, to_address, err_info, order_status, update_date, create_date  FROM nft_info where order_id = ?"
	rows, err := c.SqlDB.Query(query, OrderId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		nft := &utils.NFTInfo{}
		err := rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &nft.TickId, &nft.Total, &nft.Model, &nft.Prompt, &nft.Image, &nft.FeeTxHash, &nft.FeeTxIndex, &nft.FeeBlockHash, &nft.FeeBlockNumber, &nft.NftTxHash, &nft.NftBlockHash, &nft.NftBlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.ToAddress, &nft.ErrInfo, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)
		if err != nil {
			return nil, err
		}
		return nft, err
	}
	return nil, nil
}

func (c *DBClient) FindNftInfo(orderId, op, holder_address string, limit, offset int64) ([]*utils.NFTInfo, int64, error) {
	query := "SELECT  order_id, op, tick, tick_id, total, model, prompt, image, fee_tx_hash, fee_tx_index, fee_block_hash, fee_block_number, nft_tx_hash, nft_block_hash, nft_block_number, fee_address,to_address, holder_address, err_info, order_status,update_date, create_date  FROM nft_info  "

	where := "where"
	whereAges := []any{}
	if orderId != "" {
		where += " order_id = ? "
		whereAges = append(whereAges, orderId)
	}

	if op != "" {
		if where != "where" {
			where += " and "
		}
		where += "  op = ? "
		whereAges = append(whereAges, op)
	}

	if holder_address != "" {
		if where != "where" {
			where += " and "
		}
		where += "  (holder_address = ? or to_address = ?) "
		whereAges = append(whereAges, holder_address)
		whereAges = append(whereAges, holder_address)
	}

	order := " order by create_date desc "
	lim := " LIMIT ? OFFSET ?"
	whereAges1 := append(whereAges, limit)
	whereAges1 = append(whereAges1, offset)

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	nfts := make([]*utils.NFTInfo, 0)
	for rows.Next() {
		nft := &utils.NFTInfo{}
		var amt string
		rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &nft.TickId, &amt, &nft.Model, &nft.Prompt, &nft.Image, &nft.FeeTxHash, &nft.FeeTxIndex, &nft.FeeBlockHash, &nft.FeeBlockNumber, &nft.NftTxHash, &nft.NftBlockHash, &nft.NftBlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.ToAddress, &nft.ErrInfo, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)

		if err != nil {
			return nil, 0, err
		}

		nfts = append(nfts, nft)
	}

	query1 := "SELECT count(order_id)  FROM nft_info "
	rows1, err := c.SqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return nfts, total, nil
}

func (c *DBClient) UpdateNftInfoErr(orderId, errInfo string) error {
	query := "update nft_info set err_info = ?, order_status = 1  where order_id = ?"
	_, err := c.SqlDB.Exec(query, errInfo, orderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) FindNftCollectByTick(tick string) (*utils.NftCollect, error) {
	query := "SELECT tick_sum, total FROM nft_collect WHERE tick = ?"
	rows, err := c.SqlDB.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {

		nftc := &utils.NftCollect{}
		err := rows.Scan(&nftc.TickSum, &nftc.Total)
		if err != nil {
			return nil, err
		}

		return nftc, nil
	}

	return nil, ErrNotFound
}

func (c *DBClient) FindNftCollectAllByTick(tick string) (*utils.NftCollect, error) {
	query := "SELECT tick, tick_sum, total, model, prompt, image, create_date, holder_address, deploy_hash FROM nft_collect WHERE tick = ?"
	rows, err := c.SqlDB.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		nftc := &utils.NftCollect{}
		err := rows.Scan(&nftc.Tick, &nftc.TickSum, &nftc.Total, &nftc.Model, &nftc.Prompt, &nftc.Image, &nftc.CreateDate, &nftc.HolderAddress, &nftc.DeployHash)
		if err != nil {
			return nil, err
		}
		return nftc, nil
	}

	return nil, ErrNotFound
}

func (c *DBClient) FindNftCollectAllByTickAndId(tick string, tickId int64) (*utils.NftCollectAddress, error) {
	query := `SELECT nca.tick,
					   nca.prompt,
					   nca.image,
					   nca.create_date,
					   nca.holder_address,
					   nca.deploy_hash,
					   nc.prompt as nft_prompt,
						 nc.model as nft_model
				FROM nft_collect_address nca
					left join nft_collect nc on nca.tick = nc.tick
				WHERE nca.tick = ?
				  and nca.tick_id = ?`
	rows, err := c.SqlDB.Query(query, tick, tickId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		nftc := &utils.NftCollectAddress{}
		err := rows.Scan(&nftc.Tick, &nftc.Prompt, &nftc.Image, &nftc.CreateDate, &nftc.HolderAddress, &nftc.DeployHash, &nftc.NftPrompt, &nftc.NftModel)
		if err != nil {
			return nil, err
		}
		return nftc, nil
	}

	return nil, ErrNotFound
}

func (c *DBClient) FindNftHoldersByTick(tick string, limit, offset int64) ([]*utils.NftCollectAddress, int64, error) {
	query := "SELECT tick, tick_id, prompt, image, create_date, holder_address FROM nft_collect_address WHERE tick = ?  LIMIT ? OFFSET ? ;"
	rows, err := c.SqlDB.Query(query, tick, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*utils.NftCollectAddress
	for rows.Next() {
		result := &utils.NftCollectAddress{}
		err := rows.Scan(&result.Tick, &result.TickId, &result.Prompt, &result.Image, &result.CreateDate, &result.HolderAddress)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, result)
	}

	query1 := "SELECT count(holder_address) FROM nft_collect_address WHERE tick = ?"
	rows1, err := c.SqlDB.Query(query1, tick)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return results, total, nil
}

func (c *DBClient) FindNftCollectAddressByTickAndId(tx *sql.Tx, tick string, address string, tickId int64) (*utils.NftCollectAddress, error) {
	query := "SELECT tick_id  FROM nft_collect_address WHERE tick = ? and holder_address = ? and tick_id = ?"
	rows, err := tx.Query(query, tick, address, tickId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		nftca := &utils.NftCollectAddress{}
		err := rows.Scan(&nftca.TickId)
		if err != nil {
			return nil, err
		}
		return nftca, nil
	}

	return nil, ErrNotFound
}

func (c *DBClient) FindNftAll() ([]*NftCollect, int64, error) {
	query := `SELECT ci.tick,
				   ci.tick_sum,
				   ci.total,
				   ci.prompt,
				   ci.image,
				   ci.transactions,
				   COUNT(di.tick),
				   ci.create_date AS DeployTime,
				   ci.deploy_hash,
				   ci.introduction,
				   ci.is_check
			FROM nft_collect AS ci
					 LEFT JOIN nft_collect_address AS di ON ci.tick = di.tick
			GROUP BY ci.tick
			ORDER BY DeployTime DESC`

	rows, err := c.SqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*NftCollect
	for rows.Next() {

		result := &NftCollect{}
		err := rows.Scan(&result.Tick, &result.TickSum, &result.Total, &result.Prompt, &result.Image, &result.Transactions, &result.Holders, &result.CreateDate, &result.DeployHash, &result.Introduction, &result.IsCheck)
		if err != nil {
			return nil, 0, err
		}

		if result.IsCheck == 0 {
			de := ""
			result.Introduction = &de
			result.WhitePaper = &de
			result.Official = &de
			result.Telegram = &de
			result.Discorad = &de
			result.Twitter = &de
			result.Facebook = &de
			result.Github = &de
		}

		results = append(results, result)
	}

	query1 := "SELECT COUNT(tick) FROM nft_collect "

	rows1, err := c.SqlDB.Query(query1)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}
	return results, total, nil
}

func (c *DBClient) FindNftByAddressTick(holder_address, tick string, limit, offset int64) ([]*NftCollectAddress, int64, error) {
	query := " SELECT tick, tick_id, image, holder_address, update_date, create_date FROM nft_collect_address "

	where := "where"
	whereAges := []any{}
	if holder_address != "" {
		where += " holder_address = ? "
		whereAges = append(whereAges, holder_address)
	}

	if tick != "" {
		if where != "where" {
			where += " and "
		}
		where += "  tick = ? "
		whereAges = append(whereAges, tick)
	}

	order := " order by update_date desc "
	lim := " LIMIT ? OFFSET ? "

	whereAgesLim := append(whereAges, limit)
	whereAgesLim = append(whereAgesLim, offset)

	rows, err := c.SqlDB.Query(query+where+order+lim, whereAgesLim...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	ncas := make([]*NftCollectAddress, 0)
	for rows.Next() {
		result := &NftCollectAddress{}
		err := rows.Scan(&result.Tick, &result.TickId, &result.Image, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ncas = append(ncas, result)
	}

	query1 := " SELECT count(id)  FROM nft_collect_address "
	rows1, err := c.SqlDB.Query(query1+where, whereAges...)
	if err != nil {
		return nil, 0, err
	}

	defer rows1.Close()
	total := int64(0)
	if rows1.Next() {
		rows1.Scan(&total)
	}

	return ncas, total, nil
}

func (c *DBClient) FindNftRevertByNumber(height int64) ([]*utils.NftRevert, error) {
	query := "SELECT tick, from_address, to_address, tick_id, prompt, image, deploy_hash FROM nft_revert where block_number > ? order by id desc "
	rows, err := c.SqlDB.Query(query, height)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	reverts := make([]*utils.NftRevert, 0)
	for rows.Next() {
		revert := &utils.NftRevert{}
		err := rows.Scan(&revert.Tick, &revert.FromAddress, &revert.ToAddress, &revert.TickId, &revert.Prompt, &revert.Image, &revert.DeployHash)
		if err != nil {
			return nil, err
		}

		reverts = append(reverts, revert)
	}

	return reverts, nil
}

func (c *DBClient) DelNftRevert(tx *sql.Tx, height int64) error {
	query := "delete from nft_revert where block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}
