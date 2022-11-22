/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/wangyi/MgHash/dao/mysql"
	"github.com/wangyi/MgHash/model"
	"github.com/wangyi/MgHash/util"
	"net/http"
	"strconv"
)

func GetRecord(c *gin.Context) {
	action := c.Query("action")
	if action == "GET" {
		//获取 玩法管理
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		var total int = 0
		Db := mysql.DB
		vipEarnings := make([]model.TransactionRecordModel, 0)
		if status, isExist := c.GetQuery("status"); isExist == true {
			status, _ := strconv.Atoi(status)
			Db = Db.Where("status=?", status)
		}
		//根据时间来查询 block_timestamp
		if start, isExist := c.GetQuery("start"); isExist == true {
			if end, isExist1 := c.GetQuery("end"); isExist1 == true {
				Db = Db.Where("block_timestamp >= ? AND block_timestamp <= ?", start, end)
			}
		}
		if status, isExist := c.GetQuery("form"); isExist == true { //玩家查询
			Db = Db.Where("form=?", status)
		}

		if status, isExist := c.GetQuery("Symbol"); isExist == true { //玩家查询
			Db = Db.Where("symbol=?", status)
		}


		if PlayerMethodModelId, isExist := c.GetQuery("PlayerMethodModelId"); isExist == true { //获取玩法id  分类
			PlayerMethodModelIdId, _ := strconv.Atoi(PlayerMethodModelId)
			player := model.PlayerMethodModel{}
			err2 := mysql.DB.Where("id=?", PlayerMethodModelIdId).First(&player).Error
			if err2 != nil {
				util.JsonWrite(c, -101, nil, "非法参数")
				return
			}
			Db = Db.Where("to_address=?", player.HandAccountAddress)
		}

		Db.Table("transaction_record_models").Count(&total)
		Db = Db.Model(&vipEarnings).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
		if err := Db.Find(&vipEarnings).Error; err != nil {
			util.JsonWrite(c, -101, nil, err.Error())
			return
		}
		for k, v := range vipEarnings {
			player := model.PlayerMethodModel{}
			err := mysql.DB.Where("id=?", v.PlayerMethodModelId).First(&player).Error
			if err == nil {
				vipEarnings[k].PlayerRemark = player.Remark
				if v.BlockHash != "" {
					vipEarnings[k].BlockHash = v.BlockHash[:5] + "...." + v.BlockHash[len(v.BlockHash)-5:]
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"code":   1,
			"count":  total,
			"result": vipEarnings,
		})
		return
	}

}
