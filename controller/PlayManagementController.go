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
	"time"
)

/**
  玩法管理  添加删除 修改
*/
func SetPlay(c *gin.Context) {
	action := c.Query("action")
	if action == "GET" {
		//获取 玩法管理
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		var total int = 0
		Db := mysql.DB
		vipEarnings := make([]model.PlayerMethodModel, 0)
		if status, isExist := c.GetQuery("status"); isExist == true {
			status, _ := strconv.Atoi(status)
			Db = Db.Where("status=?", status)
		}
		Db.Table("player_method_models").Count(&total)
		Db = Db.Model(&vipEarnings).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		if err := Db.Find(&vipEarnings).Error; err != nil {
			util.JsonWrite(c, -101, nil, err.Error())
			return
		}
		//fmt.Println(vipEarnings)
		c.JSON(http.StatusOK, gin.H{
			"code":   1,
			"count":  total,
			"result": vipEarnings,
		})
		return
	}
	if action == "ADD" {
		name := c.Query("Name")
		remark := c.Query("Remark")
		HandAccountAddress := c.Query("HandAccountAddress")
		MinBetMoneyForUsdt := c.Query("MinBetMoneyForUsdt")
		ManBetMoneyForUsdt := c.Query("ManBetMoneyForUsdt")
		MinBetMoneyForXrt := c.Query("MinBetMoneyForXrt")
		ManBetMoneyForXrt := c.Query("ManBetMoneyForXrt")
		LossPerCent := c.Query("LossPerCent")
		kinds := c.Query("kinds")
		MinBetMoneyForUsdt64, _ := strconv.ParseFloat(MinBetMoneyForUsdt, 64)
		ManBetMoneyForUsdt64, _ := strconv.ParseFloat(ManBetMoneyForUsdt, 64)
		MinBetMoneyForXrt64, _ := strconv.ParseFloat(MinBetMoneyForXrt, 64)
		ManBetMoneyForXrt64, _ := strconv.ParseFloat(ManBetMoneyForXrt, 64)
		LossPerCent64, _ := strconv.ParseFloat(LossPerCent, 64)

		kind, _ := strconv.Atoi(kinds)
		add := model.PlayerMethodModel{
			Name:               name,
			Remark:             remark,
			HandAccountAddress: HandAccountAddress,
			MinBetMoneyForUsdt: MinBetMoneyForUsdt64,
			ManBetMoneyForUsdt: ManBetMoneyForUsdt64,
			MinBetMoneyForXrt:  MinBetMoneyForXrt64,
			ManBetMoneyForXrt:  ManBetMoneyForXrt64,
			Status:             2,
			Created:            time.Now().Unix(),
			LossPerCent:        LossPerCent64,
			Kinds:              kind,
		}

		//判断是否存在这个地址
		err := mysql.DB.Where("hand_account_address=?", HandAccountAddress).First(&model.PlayerMethodModel{}).Error
		if err == nil {
			util.JsonWrite(c, -101, nil, "不要重复添加")
			return
		}

		mysql.DB.Save(&add)
		util.JsonWrite(c, 200, nil, "添加成功")
	}
	if action == "UPDATE" {

		id := c.Query("ID")
		up := make(map[string]interface{})
		if HandAccountAddress, isExist := c.GetQuery("Remark"); isExist == true {
			up["remark"] = HandAccountAddress
		}

		if HandAccountAddress, isExist := c.GetQuery("Name"); isExist == true {
			up["name"] = HandAccountAddress

		}

		if HandAccountAddress, isExist := c.GetQuery("LossPerCent"); isExist == true {
			HandAccountAddress64, _ := strconv.ParseFloat(HandAccountAddress, 64)
			up["loss_per_cent"] = HandAccountAddress64
		}

		if HandAccountAddress, isExist := c.GetQuery("HandAccountAddress"); isExist == true {
			up["hand_account_address"] = HandAccountAddress
		}

		if MinBetMoneyForUsdt, isExist := c.GetQuery("MinBetMoneyForUsdt"); isExist == true {
			MinBetMoneyForUsdt64, _ := strconv.ParseFloat(MinBetMoneyForUsdt, 64)
			up["min_bet_money_for_usdt"] = MinBetMoneyForUsdt64
		}

		if MinBetMoneyForUsdt, isExist := c.GetQuery("ManBetMoneyForUsdt"); isExist == true {
			MinBetMoneyForUsdt64, _ := strconv.ParseFloat(MinBetMoneyForUsdt, 64)
			up["man_bet_money_for_usdt"] = MinBetMoneyForUsdt64
		}

		if MinBetMoneyForUsdt, isExist := c.GetQuery("MinBetMoneyForXrt"); isExist == true {
			MinBetMoneyForUsdt64, _ := strconv.ParseFloat(MinBetMoneyForUsdt, 64)
			up["min_bet_money_for_xrt"] = MinBetMoneyForUsdt64
		}
		if MinBetMoneyForUsdt, isExist := c.GetQuery("ManBetMoneyForXrt"); isExist == true {
			MinBetMoneyForUsdt64, _ := strconv.ParseFloat(MinBetMoneyForUsdt, 64)
			up["man_bet_money_for_xrt"] = MinBetMoneyForUsdt64
		}

		if MinBetMoneyForUsdt, isExist := c.GetQuery("Status"); isExist == true {
			status, _ := strconv.Atoi(MinBetMoneyForUsdt)
			up["status"] = status
		}

		if MinBetMoneyForUsdt, isExist := c.GetQuery("Kinds"); isExist == true {
			status, _ := strconv.Atoi(MinBetMoneyForUsdt)
			up["kinds"] = status
		}

		//判断是否存在这个数据
		err := mysql.DB.Where("id=?", id).First(&model.PlayerMethodModel{}).Error
		if err != nil {
			util.JsonWrite(c, -101, nil, "非法请求")
			return
		}

		err = mysql.DB.Model(&model.PlayerMethodModel{}).Where("id=?", id).Update(up).Error
		if err != nil {
			util.JsonWrite(c, -101, nil, err.Error())
			return
		}

		util.JsonWrite(c, 200, nil, "修改成功")
		return
	}

}

//获取玩家
func GetUsers(c *gin.Context) {
	action := c.Query("action")
	if action == "GET" {
		//获取 玩法管理
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		var total int = 0
		Db := mysql.DB
		vipEarnings := make([]model.UserModel, 0)
		if status, isExist := c.GetQuery("status"); isExist == true {
			status, _ := strconv.Atoi(status)
			Db = Db.Where("status=?", status)
		}

		if status, isExist := c.GetQuery("trc20_address"); isExist == true {
			//地址搜索
			Db = Db.Where("trc20_address=?", status)
		}

		//代理 搜索  一级
		if status, isExist := c.GetQuery("superior_id"); isExist == true {
			status, _ := strconv.Atoi(status)
			Db = Db.Where("superior_id=?", status)
		}
		//代理 搜索  二级
		if status, isExist := c.GetQuery("next_superior_id"); isExist == true {
			status, _ := strconv.Atoi(status)
			Db = Db.Where("next_superior_id=?", status)
		}

		//代理 搜索  三级级
		if status, isExist := c.GetQuery("next_next_superior_id"); isExist == true {
			status, _ := strconv.Atoi(status)
			Db = Db.Where("next_next_superior_id=?", status)
		}

		Db.Table("user_models").Count(&total)
		Db = Db.Model(&vipEarnings).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		if err := Db.Find(&vipEarnings).Error; err != nil {
			util.JsonWrite(c, -101, nil, err.Error())
			return
		}

		type Toll struct {
			AllBetMoneyUSDT          float64 `json:"all_bet_money_usdt"`
			AllBackMoneyUSDT         float64 `json:"all_back_money_usdt"`
			AllBetMoneyTRX           float64 `json:"all_bet_money_trx"`
			AllBackMoneyTRX          float64 `json:"all_back_money_trx"`
			SubordinateBetMoneyUSDT  float64 `json:"subordinate_bet_money_usdt"`
			SubordinateBackMoneyUSDT float64 `json:"subordinate_back_money_usdt"`
			SubordinateBetMoneyTRX   float64 `json:"subordinate_bet_money_trx"`
			SubordinateBackMoneyTRx  float64 `json:"subordinate_back_money_t_rx"`
		}
		for k, v := range vipEarnings {
			//所有
			a := make([]model.TransactionRecordModel, 0)
			mysql.DB.Raw("SELECT  * FROM transaction_record_models WHERE form=?   GROUP BY   bet_name", v.Trc20Address).Scan(&a)
			if len(a) > 0 {
				for _, l := range a {
					vipEarnings[k].VipTotal.All = append(vipEarnings[k].VipTotal.All, l.BetName)
				}
			}
			//当日的
			b := make([]model.TransactionRecordModel, 0)
			mysql.DB.Raw("SELECT  * FROM transaction_record_models WHERE form=? And  date =? GROUP BY   bet_name", v.Trc20Address, time.Now().Format("2006-01-02")).Scan(&b)
			if len(b) > 0 {
				for _, l := range b {
					vipEarnings[k].VipTotal.Today = append(vipEarnings[k].VipTotal.Today, l.BetName)
				}
			}
			// 获取这个号的总投注  和  返回金额
			var p Toll
			mysql.DB.Raw("SELECT SUM(money) as  all_bet_money_usdt ,SUM(back_money)  as all_back_money_usdt FROM transaction_record_models WHERE form=? AND  status !=3 AND  (symbol=?  OR symbol=?)", v.Trc20Address, "USDT", "usdt").Scan(&p)
			mysql.DB.Raw("SELECT SUM(money) as  all_bet_money_trx ,SUM(back_money)  as all_back_money_trx FROM transaction_record_models WHERE form=? AND  status !=3 AND  (symbol=?  OR symbol=?)", v.Trc20Address, "TRX", "trx").Scan(&p)
			vipEarnings[k].AllBetMoneyUSDT = p.AllBetMoneyUSDT
			vipEarnings[k].AllBackMoneyUSDT = p.AllBackMoneyUSDT
			vipEarnings[k].AllBetMoneyTRX = p.AllBetMoneyTRX
			vipEarnings[k].AllBackMoneyTRX = p.AllBackMoneyTRX
			mysql.DB.Raw("SELECT SUM(money) as  all_bet_money_usdt ,SUM(back_money)  as all_back_money_usdt FROM transaction_record_models WHERE form=? AND  status !=3 AND  (symbol=?  OR symbol=? )", v.Trc20Address, "USDT", "usdt").Scan(&p)
			o := make([]model.UserModel, 0)
			mysql.DB.Where("superior_id=?", v.ID).Find(&o)
			if len(o) > 0 {
				for _, i2 := range o {
					mysql.DB.Raw("SELECT SUM(money) as  subordinate_bet_money_usdt ,SUM(back_money)  as subordinate_back_money_usdt FROM transaction_record_models WHERE form=? AND  status !=3 AND  (symbol=?  OR symbol=?)", i2.Trc20Address, "USDT", "usdt").Scan(&p)
					mysql.DB.Raw("SELECT SUM(money) as  subordinate_bet_money_trx ,SUM(back_money)  as subordinate_back_money_t_rx FROM transaction_record_models WHERE form=? AND  status !=3 AND  (symbol=?  OR symbol=?)", i2.Trc20Address, "USDT", "usdt").Scan(&p)
					vipEarnings[k].SubordinateBetMoneyUSDT = vipEarnings[k].SubordinateBetMoneyUSDT + p.SubordinateBetMoneyUSDT
					vipEarnings[k].SubordinateBetMoneyTRX = vipEarnings[k].SubordinateBetMoneyTRX + p.SubordinateBetMoneyTRX
					vipEarnings[k].SubordinateBackMoneyUSDT = vipEarnings[k].SubordinateBackMoneyUSDT + p.SubordinateBackMoneyUSDT
					vipEarnings[k].SubordinateBackMoneyTRx = vipEarnings[k].SubordinateBackMoneyTRx + p.SubordinateBackMoneyTRx
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

	if action == "UPDATE" {

		id := c.Query("id")
		ups := make(map[string]interface{})

		if remark, isExist := c.GetQuery("remark"); isExist == true {
			ups["Remark"] = remark
		}

		err := mysql.DB.Model(&model.UserModel{}).Where("id=?", id).Update(ups).Error
		if err != nil {
			util.JsonWrite(c, -101, nil, "修改失败")
			return
		}
		util.JsonWrite(c, 200, nil, "修改成功")
		return
	}

}
