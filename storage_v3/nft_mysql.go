package storage_v3

import (
	"database/sql"
	"github.com/unielon-org/unielon-indexer/models"
)

func (e *MysqlClient) InstallNftRevert(tx *sql.Tx, tick, from, to string, tickId int64, height int64, prompt, image, imagePath, deployHash string) error {
	exec := "INSERT INTO nft_revert (tick, from_address, to_address, tick_id, block_number, prompt, image, image_path, deploy_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, tick, from, to, tickId, height, prompt, image, imagePath, deployHash)
	if err != nil {
		return err
	}
	return nil
}

func (c *MysqlClient) FindNftInfoById(OrderId string) (*models.NftInfo, error) {
	query := "SELECT  order_id, op, tick, tick_id, total, model, prompt, image_path, fee_tx_hash, tx_hash, block_hash, block_number, fee_address, holder_address, to_address, err_info, order_status, update_date, create_date  FROM nft_info where order_id = ?"
	rows, err := c.MysqlDB.Query(query, OrderId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if rows.Next() {
		nft := &models.NftInfo{}
		err := rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &nft.TickId, &nft.Total, &nft.Model, &nft.Prompt, &nft.ImagePath, &nft.FeeTxHash, &nft.TxHash, &nft.BlockHash, &nft.BlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.ToAddress, &nft.ErrInfo, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)
		if err != nil {
			return nil, err
		}
		return nft, err
	}
	return nil, nil
}

func (c *MysqlClient) FindNftInfo(orderId, op, holder_address string, limit, offset int64) ([]*models.NftInfo, int64, error) {
	query := "SELECT  order_id, op, tick, tick_id, total, model, prompt, image_path, fee_tx_hash, tx_hash, block_hash, block_number, fee_address,to_address, holder_address, err_info, order_status, update_date, create_date  FROM nft_info  "

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

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAges1...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	nfts := make([]*models.NftInfo, 0)
	for rows.Next() {
		nft := &models.NftInfo{}
		var amt string
		rows.Scan(&nft.OrderId, &nft.Op, &nft.Tick, &nft.TickId, &amt, &nft.Model, &nft.Prompt, &nft.ImagePath, &nft.FeeTxHash, &nft.TxHash, &nft.BlockHash, &nft.BlockNumber, &nft.FeeAddress, &nft.HolderAddress, &nft.ToAddress, &nft.ErrInfo, &nft.OrderStatus, &nft.UpdateDate, &nft.CreateDate)

		if err != nil {
			return nil, 0, err
		}

		nfts = append(nfts, nft)
	}

	query1 := "SELECT count(order_id)  FROM nft_info "
	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
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

func (c *MysqlClient) UpdateNftInfoErr(orderId, errInfo string) error {
	query := "update nft_info set err_info = ?, order_status = 1  where order_id = ?"
	_, err := c.MysqlDB.Exec(query, errInfo, orderId)
	if err != nil {
		return err
	}
	return nil
}

func (c *MysqlClient) FindNftCollectAllByTick(tick string) (*models.NftCollect, error) {
	query := "SELECT tick, tick_sum, total, model, prompt, image_path, create_date, holder_address, deploy_hash FROM nft_collect WHERE tick = ?"
	rows, err := c.MysqlDB.Query(query, tick)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		nftc := &models.NftCollect{}
		err := rows.Scan(&nftc.Tick, &nftc.TickSum, &nftc.Total, &nftc.Model, &nftc.Prompt, &nftc.ImagePath, &nftc.CreateDate, &nftc.HolderAddress, &nftc.DeployHash)
		if err != nil {
			return nil, err
		}
		return nftc, nil
	}

	return nil, ErrNotFound
}

func (c *MysqlClient) FindNftCollectAllByTickAndId(tick string, tickId int64) (*models.NftCollectAddress, error) {
	query := `SELECT nca.tick,
					   nca.prompt,
					   nca.image_path,
					   UNIX_TIMESTAMP(nca.create_date),
					   nca.holder_address,
					   nca.deploy_hash,
					   nc.prompt as nft_prompt,
						 nc.model as nft_model,
						nca.is_check
				FROM nft_collect_address nca
					left join nft_collect nc on nca.tick = nc.tick
				WHERE nca.tick = ?
				  and nca.tick_id = ?`
	rows, err := c.MysqlDB.Query(query, tick, tickId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		nftc := &models.NftCollectAddress{}
		err := rows.Scan(&nftc.Tick, &nftc.Prompt, &nftc.ImagePath, &nftc.CreateDate, &nftc.HolderAddress, &nftc.DeployHash, &nftc.NftPrompt, &nftc.NftModel, &nftc.IsCheck)
		if err != nil {
			return nil, err
		}
		if nftc.IsCheck == 1 {
			nftc.ImagePath = ""
		}
		return nftc, nil
	}

	return nil, ErrNotFound
}

func (c *MysqlClient) FindNftHoldersByTick(tick string, limit, offset int64) ([]*models.NftCollectAddress, int64, error) {
	query := "SELECT tick, tick_id, prompt, image_path, create_date, holder_address, is_check FROM nft_collect_address WHERE tick = ?  LIMIT ? OFFSET ? ;"
	rows, err := c.MysqlDB.Query(query, tick, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*models.NftCollectAddress
	for rows.Next() {
		result := &models.NftCollectAddress{}
		err := rows.Scan(&result.Tick, &result.TickId, &result.Prompt, &result.ImagePath, &result.CreateDate, &result.HolderAddress, &result.IsCheck)
		if err != nil {
			return nil, 0, err
		}
		if result.IsCheck == 1 {
			result.ImagePath = ""
		}

		results = append(results, result)
	}

	query1 := "SELECT count(holder_address) FROM nft_collect_address WHERE tick = ?"
	rows1, err := c.MysqlDB.Query(query1, tick)
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

func (c *MysqlClient) FindNftAll() ([]*models.NftCollect, int64, error) {
	query := `SELECT ci.tick,
				   ci.tick_sum,
				   ci.total,
				   ci.prompt,
				   ci.image_path,
				   ci.transactions,
				   COUNT(di.tick),
				   UNIX_TIMESTAMP(ci.create_date) AS DeployTime,
				   ci.deploy_hash,
				   ci.introduction,
				   ci.is_check
			FROM nft_collect AS ci
					 LEFT JOIN nft_collect_address AS di ON ci.tick = di.tick
			GROUP BY ci.tick
			ORDER BY DeployTime DESC`

	rows, err := c.MysqlDB.Query(query)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	var results []*models.NftCollect
	for rows.Next() {

		result := &models.NftCollect{}
		err := rows.Scan(&result.Tick, &result.TickSum, &result.Total, &result.Prompt, &result.ImagePath, &result.Transactions, &result.Holders, &result.CreateDate, &result.DeployHash, &result.Introduction, &result.IsCheck)
		if err != nil {
			return nil, 0, err
		}

		if result.IsCheck == 1 {
			continue
		}

		results = append(results, result)
	}

	query1 := "SELECT COUNT(tick) FROM nft_collect "

	rows1, err := c.MysqlDB.Query(query1)
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

func (c *MysqlClient) FindNftByAddressTick(holder_address, tick string, limit, offset int64) ([]*models.NftCollectAddress, int64, error) {
	query := " SELECT tick, tick_id, image_path, holder_address,update_date, create_date FROM nft_collect_address "

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

	rows, err := c.MysqlDB.Query(query+where+order+lim, whereAgesLim...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()
	ncas := make([]*models.NftCollectAddress, 0)
	for rows.Next() {
		result := &models.NftCollectAddress{}
		err := rows.Scan(&result.Tick, &result.TickId, &result.ImagePath, &result.HolderAddress, &result.UpdateDate, &result.CreateDate)
		if err != nil {
			return nil, 0, err
		}
		ncas = append(ncas, result)
	}

	query1 := " SELECT count(id)  FROM nft_collect_address "
	rows1, err := c.MysqlDB.Query(query1+where, whereAges...)
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
