package router

import (
	"github.com/dogecoinw/doged/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unielon-org/unielon-indexer/models"
	"github.com/unielon-org/unielon-indexer/storage"
	"github.com/unielon-org/unielon-indexer/utils"
	"github.com/unielon-org/unielon-indexer/verifys"
	"net/http"
)

type FileRouter struct {
	dbc  *storage.DBClient
	node *rpcclient.Client
	ipfs *shell.Shell

	verify *verifys.Verifys
}

func NewFileRouter(db *storage.DBClient, node *rpcclient.Client, ipfs *shell.Shell, verify *verifys.Verifys) *FileRouter {
	return &FileRouter{
		dbc:    db,
		node:   node,
		ipfs:   ipfs,
		verify: verify,
	}
}

func (r *FileRouter) Order(c *gin.Context) {
	params := &struct {
		OrderId       string `json:"order_id"`
		Op            string `json:"op"`
		HolderAddress string `json:"holder_address"`
		ToAddress     string `json:"to_address"`
		BlockNumber   int64  `json:"block_number"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	filter := &models.FileInfo{
		OrderId:       params.OrderId,
		Op:            params.Op,
		HolderAddress: params.HolderAddress,
		ToAddress:     params.ToAddress,
		BlockNumber:   params.BlockNumber,
	}

	var nfts []*models.FileInfo
	var total int64
	err := r.dbc.DB.Model(&models.FileInfo{}).
		Where(filter).
		Count(&total).
		Limit(params.Limit).
		Offset(params.OffSet).
		Order("create_date desc").
		Find(&nfts).Error
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "server error"
		c.JSON(http.StatusOK, result)
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = nfts
	result.Total = total

	c.JSON(http.StatusOK, result)

}

func (r *FileRouter) CollectAddress(c *gin.Context) {
	params := &struct {
		FileId        string `json:"file_id"`
		HolderAddress string `json:"holder_address"`
		NoMeta        int    `json:"no_meta"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	type QueryResult struct {
		FileId        string           `gorm:"column:file_id" json:"file_id"`
		FilePath      string           `gorm:"column:file_path" json:"file_path"`
		FileLength    int              `gorm:"column:file_length" json:"file_length"`
		FileType      string           `gorm:"column:file_type" json:"file_type"`
		HolderAddress string           `gorm:"column:holder_address" json:"holder_address"`
		UpdateDate    models.LocalTime `gorm:"column:update_date" json:"update_date"`
		CreateDate    models.LocalTime `gorm:"column:create_date" json:"create_date"`
		MetaId        string           `gorm:"column:meta_id" json:"meta_id"`
		MetaName      string           `gorm:"column:meta_name" json:"meta_name"`
		FileName      string           `gorm:"column:file_name" json:"file_name"`
		ExId          string           `gorm:"column:ex_id" json:"ex_id"`
		Tick          string           `gorm:"column:tick" json:"tick"`
		Amt           models.Number    `gorm:"column:amt" json:"amt"`
	}

	var results []QueryResult

	var err error
	var total int64

	subQuery := r.dbc.DB.Table("file_collect_address").
		Select("file_collect_address.file_id, file_collect_address.file_path, file_collect_address.file_length, file_collect_address.file_type, file_collect_address.holder_address, file_collect_address.update_date, file_collect_address.create_date, file_meta_inscription.meta_id, file_meta_inscription.name").
		Joins("left join file_meta_inscription on file_collect_address.file_id = file_meta_inscription.file_id")

	if params.HolderAddress != "" {
		subQuery.Where("file_collect_address.holder_address = ?", params.HolderAddress)
	}

	if params.FileId != "" {

		subQuery = r.dbc.DB.Table("file_collect_address").
			Select("file_collect_address.file_id, file_collect_address.file_path,file_collect_address.file_length, file_collect_address.file_type, file_collect_address.holder_address, file_collect_address.update_date, file_collect_address.create_date, file_meta_inscription.meta_id, file_meta_inscription.name as file_name, file_meta.name as meta_name, fec.ex_id, fec.tick, fec.amt").
			Joins("left join file_meta_inscription on file_collect_address.file_id = file_meta_inscription.file_id").
			Joins("left join file_meta on file_meta.meta_id = file_meta_inscription.meta_id").
			Joins("left join (select ex_id, tick, file_id, amt from file_exchange_collect where amt != amt_finish) as fec on file_collect_address.file_id = fec.file_id")

		subQuery.Where("file_collect_address.file_id = ?", params.FileId)
		err = subQuery.Group("file_collect_address.file_id, file_collect_address.file_path, file_collect_address.holder_address, file_collect_address.update_date, file_collect_address.create_date, file_meta_inscription.meta_id, file_meta_inscription.name, file_exchange_collect.ex_id, file_exchange_collect.tick, file_exchange_collect.amt, meta_name, file_collect_address.file_length, file_collect_address.file_type").
			Find(&results).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		result := &utils.HttpResult{}
		result.Code = 200
		result.Msg = "success"
		result.Data = results
		result.Total = total

		c.JSON(http.StatusOK, result)
		return
	}

	if params.NoMeta == 1 {
		subQuery.Where("file_meta_inscription.meta_id is null")
	} else if params.NoMeta == 2 {
		subQuery.Where("file_meta_inscription.meta_id is not null")
	}

	err = subQuery.Group("file_collect_address.file_id, file_collect_address.file_path, file_collect_address.file_length, file_collect_address.file_type, file_collect_address.holder_address, file_collect_address.update_date, file_collect_address.create_date, file_meta_inscription.meta_id, file_meta_inscription.name").
		Count(&total).
		Limit(params.Limit).
		Offset(params.OffSet).
		Find(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *FileRouter) CollectionsInscriptions(c *gin.Context) {

	params := &struct {
		MetaId string `json:"meta_id"`
		Limit  int    `json:"limit"`
		OffSet int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	names := make([]string, 0)
	err := r.dbc.DB.Model(&models.FileMetaAttribute{}).Select("name").Where("meta_id = ?", params.MetaId).Limit(params.Limit).Offset(params.OffSet).Find(&names).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	infos := make([]models.FileMetaAttribute, 0)
	err = r.dbc.DB.Where("name in ?", names).Find(&infos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	type attribute struct {
		TraitType string `json:"trait_type"`
		Value     string `json:"value"`
	}

	type mate struct {
		Name       string      `json:"name"`
		Attributes []attribute `json:"attributes"`
	}

	type resultStruct struct {
		ID   string `json:"id"`
		Meta mate   `json:"meta"`
	}

	resultDatas := make([]resultStruct, 0)

	name := ""
	attributes := make([]attribute, 0)
	resultData := resultStruct{}

	for _, item := range infos {

		if name != item.Name {
			metaData := mate{
				Name:       item.Name,
				Attributes: attributes,
			}

			resultData.ID = item.FileId
			resultData.Meta = metaData
			resultDatas = append(resultDatas, resultData)

			name = item.Name
			resultData = resultStruct{}
			attributes = make([]attribute, 0)
		}

		attributes = append(attributes, attribute{
			TraitType: item.TraitType,
			Value:     item.Value,
		})
	}

	resultr := &utils.HttpResult{}
	resultr.Code = 200
	resultr.Msg = "success"
	resultr.Data = resultDatas

	c.JSON(http.StatusOK, resultr)
}

func (r *FileRouter) Collections(c *gin.Context) {

	params := &struct {
		MetaId        string `json:"meta_id"`
		HolderAddress string `json:"holder_address"`
		Limit         int    `json:"limit"`
		OffSet        int    `json:"offset"`
	}{
		Limit:  10,
		OffSet: 0,
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	filter := &models.FileMeta{
		MetaId:  params.MetaId,
		IsCheck: 1,
	}

	if params.HolderAddress != "" {
		filter.MetaId = ""
		filter.HolderAddress = params.HolderAddress
		filter.IsCheck = 0
	}

	infos := make([]models.FileMeta, 0)
	total := int64(0)
	err := r.dbc.DB.Model(&models.FileMeta{}).Where(filter).Count(&total).Limit(params.Limit).Offset(params.OffSet).Find(&infos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = infos
	result.Total = total

	c.JSON(http.StatusOK, result)
}

func (r *FileRouter) CollectionsAttributes(c *gin.Context) {

	params := &struct {
		MetaId string `json:"meta_id"`
		FileId string `json:"file_id"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	subresults := make([]struct {
		TraitType string `json:"trait_type"`
		Value     string `json:"value"`
		Count     int    `gorm:"column:count_"  json:"count"`
	}, 0)

	subQuery := r.dbc.DB.Model(&models.FileMetaAttribute{}).
		Select("trait_type, value, count(value) as count_")

	if params.FileId != "" {
		subQuery.Where("file_id = ?", params.FileId)
	}
	if params.MetaId != "" {
		subQuery.Where("meta_id = ?", params.MetaId)
	}

	err := subQuery.Group("trait_type, value").
		Find(&subresults).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	results := make(map[string]map[string]int)
	for _, item := range subresults {
		if results[item.TraitType] == nil {
			results[item.TraitType] = make(map[string]int)
		}
		results[item.TraitType][item.Value] = item.Count
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"
	result.Data = results

	c.JSON(http.StatusOK, result)

}

func (r *FileRouter) UploadMeta(c *gin.Context) {
	params := &struct {
		MetaId          string `json:"meta_id"`
		InscriptionIcon string `json:"inscription_icon"`
		SigMsg          string `json:"sig_msg"`
		Description     string `json:"description"`
		DiscordLink     string `json:"discord_link"`
		Icon            string `json:"icon"`
		Name            string `json:"name"`
		Slug            string `json:"slug"`
		TwitterLink     string `json:"twitter_link"`
		WebsiteLink     string `json:"website_link"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	inAddress, err := utils.GetAddressFromSig(params.Name, params.SigMsg)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	if params.MetaId != "" {

		fileMeta := &models.FileMeta{}
		err = r.dbc.DB.Where("meta_id = ?", params.MetaId).First(fileMeta).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		if fileMeta.HolderAddress != inAddress {
			result := &utils.HttpResult{}
			result.Code = 500
			result.Msg = "address not match"
			c.JSON(http.StatusOK, result)
			return
		}

		fileMeta.Description = params.Description
		fileMeta.DiscordLink = params.DiscordLink
		fileMeta.Icon = params.Icon
		fileMeta.Name = params.Name
		fileMeta.Slug = params.Slug
		fileMeta.TwitterLink = params.TwitterLink
		fileMeta.WebsiteLink = params.WebsiteLink

		err = r.dbc.DB.Save(&fileMeta).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		result := &utils.HttpResult{}
		result.Code = 200
		result.Msg = "success"
		result.Data = fileMeta
		c.JSON(http.StatusOK, result)
		return

	} else {
		metaId := uuid.New()

		fileMeta := &models.FileMeta{
			MetaId:        metaId.String(),
			Description:   params.Description,
			DiscordLink:   params.DiscordLink,
			Icon:          params.Icon,
			Name:          params.Name,
			Slug:          params.Slug,
			TwitterLink:   params.TwitterLink,
			WebsiteLink:   params.WebsiteLink,
			HolderAddress: inAddress,
		}

		err = r.dbc.DB.Save(&fileMeta).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		result := &utils.HttpResult{}
		result.Code = 200
		result.Msg = "success"
		result.Data = fileMeta
		c.JSON(http.StatusOK, result)
	}

	return
}

func (r *FileRouter) UploadInscriptionsMeta(c *gin.Context) {
	params := &struct {
		SigMsg   string `json:"sig_msg"`
		MetaId   string `json:"meta_id"`
		MetaName string `json:"meta_name"`
		Metas    []struct {
			Id   string `json:"id"`
			Meta struct {
				Name       string `json:"name"`
				Attributes []struct {
					TraitType string `json:"trait_type"`
					Value     string `json:"value"`
				} `json:"attributes"`
			} `json:"meta"`
		} `json:"metas"`
	}{}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	inAddress, err := utils.GetAddressFromSig(params.MetaName, params.SigMsg)
	if err != nil {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = err.Error()
		c.JSON(http.StatusOK, result)
		return
	}

	fileMeta := &models.FileMeta{}
	err = r.dbc.DB.Where("meta_id = ?", params.MetaId).First(fileMeta).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if fileMeta.HolderAddress != inAddress {
		result := &utils.HttpResult{}
		result.Code = 500
		result.Msg = "address not match"
		c.JSON(http.StatusOK, result)
		return
	}

	for _, item := range params.Metas {

		fileInscription := &models.FileMetaInscription{
			FileId: item.Id,
			MetaId: params.MetaId,
			Name:   item.Meta.Name,
		}

		err = r.dbc.DB.Create(&fileInscription).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// 删除之前对于图片的描述
		err = r.dbc.DB.Where("name = ? and meta_id = ?", item.Meta.Name, params.MetaId).Delete(&models.FileMetaAttribute{}).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		fileAttribute := &models.FileMetaAttribute{
			FileId: item.Id,
			MetaId: params.MetaId,
			Name:   item.Meta.Name,
		}

		if len(item.Meta.Attributes) == 0 {
			err = r.dbc.DB.Create(&fileAttribute).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			for _, attr := range item.Meta.Attributes {

				fileAttribute = &models.FileMetaAttribute{
					FileId:    item.Id,
					MetaId:    params.MetaId,
					Name:      item.Meta.Name,
					TraitType: attr.TraitType,
					Value:     attr.Value,
				}

				err = r.dbc.DB.Create(&fileAttribute).Error
				if err != nil {
					c.JSON(http.StatusInternalServerError, err.Error())
					return
				}
			}
		}
	}

	result := &utils.HttpResult{}
	result.Code = 200
	result.Msg = "success"

	c.JSON(http.StatusOK, result)
}
