/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package web_api

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/wangyi/MgHash/dao/mysql"
	"github.com/wangyi/MgHash/dao/redis"
	"github.com/wangyi/MgHash/model"
	"github.com/wangyi/MgHash/tools"
	"github.com/wangyi/MgHash/util"
	"net/http"
	"strconv"
	"time"
)

//注册
func Register(c *gin.Context) {
	Trc20 := c.Query("Trc20")
	lang := c.Query("lang")
	if len(Trc20) != 34 {
		if lang == "zh_CN" {
			util.ReturnError101(c, "请输出正确的地址")
		} else {
			util.ReturnError101(c, "Please enter the correct address")
		}
		return
	}
	use := model.UserModel{Trc20Address: Trc20}
	//校验地址的合格性?
	if InvitationCode, isE := c.GetQuery("InvitationCode"); isE == true {
		//判断邀请码是否存在
		user := model.UserModel{}
		err := mysql.DB.Where("invitation_code=?", InvitationCode).First(&user).Error
		if err == nil {
			use.SuperiorId = int(user.ID)
			use.NextSuperiorId = user.SuperiorId
			use.NextNextSuperiorId = user.NextSuperiorId
		}
	}
	_, err := use.AddUser(mysql.DB, redis.Rdb, lang)
	if err != nil {
		util.ReturnError101(c, err.Error())
		return
	}
	if lang == "zh_CN" {
		util.ReturnOk200Data(c, use.InvitationCode, "注册成功")
	} else {
		util.ReturnOk200Data(c, use.InvitationCode, "registered successfully")
	}
	return
}

//登录
func Login(c *gin.Context) {
	Trc20 := c.Query("Trc20")
	lang := c.Query("lang")
	user := model.UserModel{}
	err := mysql.DB.Where("trc20_address=?", Trc20).First(&user).Error
	if err != nil {
		if lang == "zh_CN" {
			util.ReturnError101(c, "请先注册")
		} else {
			util.ReturnOk200(c, "Please register first")
		}
		return
	}
	//
	util.ReturnOk200Data(c, user.InvitationCode, "OK")
	return
}

//获取配置
func GetConfig(c *gin.Context) {
	config := model.ConfigModel{}
	mysql.DB.Where("id=?", 1).First(&config)
	util.ReturnOk200Data(c, config, "ok")
	return
}

//赚币获取(累计获得佣金,一级二级三级的好友个数)
func EeaMoney(c *gin.Context) {

	action := c.Query("action")
	Trc20 := c.Query("Trc20")

	if len(Trc20) != 34 {
		util.ReturnError101(c, "error")
		return
	}
	//赚钱
	if action == "earnings" {
		//传值地址过来查询
		userData := model.UserModel{}
		err := mysql.DB.Where("trc20_address=?", Trc20).First(&userData).Error
		if err != nil {
			util.ReturnError101(c, err.Error())
			return
		}
		type EeaMoneyType struct {
			One            int     `json:"one"`
			Two            int     `json:"two"`
			Three          int     `json:"three"`
			InvitationCode string  `json:"invitation_code"`
			Commission     float64 `json:"commission"`
		}
		var ea EeaMoneyType
		ea.InvitationCode = userData.InvitationCode
		mysql.DB.Model(&model.UserModel{}).Where("superior_id=?", userData.ID).Count(&ea.One)
		mysql.DB.Model(&model.UserModel{}).Where("next_superior_id=?", userData.ID).Count(&ea.Two)
		mysql.DB.Model(&model.UserModel{}).Where("next_next_superior_id=?", userData.ID).Count(&ea.Three)
		mysql.DB.Model(model.CommissionModel{}).Where("to_address=?", Trc20).Count(&ea.Commission)
		util.ReturnOk200Data(c, ea, "ok")
		return
	}
	//list  佣金明细
	if action == "list" {
		//获取 玩法管理
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))
		var total int = 0
		Db := mysql.DB.Where("to_address=?", Trc20)
		vipEarnings := make([]model.CommissionModel, 0)
		Db.Table("commission_models").Count(&total)
		Db = Db.Model(&vipEarnings).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		if err := Db.Find(&vipEarnings).Error; err != nil {
			util.JsonWrite(c, -101, nil, err.Error())
			return
		}

		for k, v := range vipEarnings {
			vipEarnings[k].CreatedString = time.Unix(v.Created, 0).Format("2006-01-02")
			vipEarnings[k].SuccessTimeString = time.Unix(v.SuccessTime, 0).Format("2006-01-02")

		}

		c.JSON(http.StatusOK, gin.H{
			"code":   1,
			"count":  total,
			"result": vipEarnings,
		})
		return
	}

}

//投注记录
func GetBetsList(c *gin.Context) {

	action := c.Query("action")
	Trc20 := c.Query("Trc20")
	if len(Trc20) != 34 {
		util.ReturnError101(c, "error")
		return
	}
	if action == "GET" {
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))

		var total int = 0
		Db := mysql.DB.Where("form=?", Trc20)

		if BetName, isExist := c.GetQuery("BetName"); isExist == true {
			Db = Db.Where("bet_name=?", BetName)
		}

		if days, isExist := c.GetQuery("Day"); isExist == true {
			if days == "dt" {
				Db = Db.Where("date=?", time.Now().AddDate(0, 0, -1).Format("2006-01-02"))
			} else if days == "st" {
				Db = Db.Where("date=? or date =? or date =?", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), time.Now().AddDate(0, 0, -2).Format("2006-01-02"), time.Now().AddDate(0, 0, -3).Format("2006-01-02"))
			} else if days == "qt" {
				Db = Db.Where("date=? or date =? or date =? or date=? or date=? or date=? or date=?", time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
					time.Now().AddDate(0, 0, -2).Format("2006-01-02"), time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
					time.Now().AddDate(0, 0, -4).Format("2006-01-02"), time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
					time.Now().AddDate(0, 0, -6).Format("2006-01-02"), time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
			} else if days == "shit" {
				Db = Db.Where("month=?", tools.ReturnTheMonth())
			}
		}
		if status, isExist := c.GetQuery("Status"); isExist == true {
			//状态 1输  2 赢 3无效
			Db = Db.Where("status=?", status)
		}
		vipEarnings := make([]model.TransactionRecordModel, 0)
		Db.Table("transaction_record_models").Count(&total)
		Db = Db.Model(&vipEarnings).Offset((page - 1) * limit).Limit(limit).Order("created desc")
		if err := Db.Find(&vipEarnings).Error; err != nil {
			util.JsonWrite(c, -101, nil, err.Error())
			return
		}

		for k, v := range vipEarnings {

			if v.BetName == "牛牛" {
				vipEarnings[k].BetName = "A"
			}
			if v.BetName == "尾数" || v.BetName == "幸运" {
				vipEarnings[k].BetName = "B"
			}
			if v.BetName == "单双" {
				vipEarnings[k].BetName = "C"
			}
			if v.BetName == "百家乐" {
				vipEarnings[k].BetName = "D"
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

//我的邀请
func MyInvite(c *gin.Context) {

	action := c.Query("action")
	Trc20 := c.Query("Trc20")
	if len(Trc20) != 34 {
		util.ReturnError101(c, "error")
		return
	}
	if action == "GET" {
		user := model.UserModel{}
		err := mysql.DB.Where("trc20_address=?", Trc20).First(&user).Error
		if err != nil {
			util.ReturnError101(c, "error")
			return
		}
		userArray := make([]model.UserModel, 0)
		var Data []OnePeople
		//先查一级的用户
		err = mysql.DB.Where("superior_id=?", user.ID).Find(&userArray).Error
		if err == nil {
			Data = ReturnOnePeople(mysql.DB, userArray, Data, 1)
		}
		err = mysql.DB.Where("next_superior_id=?", user.ID).Find(&userArray).Error
		if err == nil {
			Data = ReturnOnePeople(mysql.DB, userArray, Data, 2)

		}
		err = mysql.DB.Where("next_next_superior_id=?", user.ID).Find(&userArray).Error
		if err == nil {
			Data = ReturnOnePeople(mysql.DB, userArray, Data, 3)
		}
		util.ReturnOk200Data(c, Data, "ok")
		return

	}

}

type OnePeople struct {
	RegisterTime       int64   `json:"register_time"`
	Trc20Address       string  `json:"trc_20_address"`
	Level              int     `json:"level"` //1 2 3
	TodayBet           float64 `json:"today_bet"`
	TodayWin           float64 `json:"today_win"`
	MonthBet           float64 `json:"month_bet"`
	MonthWin           float64 `json:"month_win"`
	RegisterTimeString string
}

func ReturnOnePeople(db *gorm.DB, userArray []model.UserModel, Data []OnePeople, level int) []OnePeople {
	for _, v := range userArray {
		o := OnePeople{}
		o.RegisterTimeString = time.Unix(v.Created, 0).Format("2006-01-02")
		o.Trc20Address = v.Trc20Address
		o.Level = level
		o.RegisterTime = v.Created //状态 1输  2 赢 3无效
		//今投注
		db.Raw("SELECT SUM(money) as today_bet  FROM transaction_record_models   WHERE form= ? AND status !=? and date=?", v.Trc20Address, 3, time.Now().Format("2006-01-02")).Scan(&o)
		//今日中奖
		db.Raw("SELECT SUM(money) as today_win  FROM transaction_record_models   WHERE form= ? AND status =? and date=?", v.Trc20Address, 2, time.Now().Format("2006-01-02")).Scan(&o)
		//本月投注
		db.Raw("SELECT SUM(money) as today_bet  FROM transaction_record_models   WHERE form= ? AND status !=? and month=?", v.Trc20Address, 3, tools.ReturnTheMonth()).Scan(&o)
		//本月中奖
		db.Raw("SELECT SUM(money) as today_win  FROM transaction_record_models   WHERE form= ? AND status =? and month=?", v.Trc20Address, 2, tools.ReturnTheMonth()).Scan(&o)
		Data = append(Data, o)
	}
	return Data
}
