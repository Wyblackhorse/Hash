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
	"github.com/wangyi/MgHash/process"
	"github.com/wangyi/MgHash/util"
	"net/http"
	"strconv"
	"time"
)

/**
  每日统计  手动执行调用
*/
func SetEverydayStatistics(c *gin.Context) {
	id := c.Query("ID")
	timeDate := time.Now().Format("2006-01-02")
	if date, isExist := c.GetQuery("date"); isExist == true {
		timeDate = date
	}

	//定时任务
	if _, isExist := c.GetQuery("zt"); isExist == true {
		go process.YesEvery()
		util.JsonWrite(c, 200, nil, "执行成功")
		return
	}

	player := model.PlayerMethodModel{}
	err := mysql.DB.Where("id=?", id).First(&player).Error
	if err != nil {
		util.JsonWrite(c, -101, nil, "非法请求")
		return
	}

	//Symbol       string  //币种 USDT  or XRT
	var BetNum int = 0
	var BetPeopleNum int = 0
	var WinNum int = 0
	var FailNum int = 0
	var NoBackNum int = 0
	type WinAccount struct {
		WinAccount  float64 `json:"total_income"`  // 中奖金额
		BackMoney   float64 `json:"back_money"`    //派奖金额
		NoBackMoney float64 `json:"no_back_money"` //未派奖金额数
		Profit      float64 `json:"profit"`        //盈利
		BetMoney    float64 `json:"bet_money"`     //下注金额
	}

	var result WinAccount

	//record := make([]model.TransactionRecordModel, 0)
	//下注笔数统计
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Count(&BetNum)
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Group("form").Count(&BetPeopleNum)
	////状态 1输  2 赢 3无效
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Count(&WinNum)
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=1").Count(&FailNum)
	//中奖金额   派奖金额
	mysql.DB.Table("transaction_record_models").Select("sum(result) as win_account ,sum(back_money)as back_money").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Scan(&result)
	//未派奖金额数 是否派发   1 已经派发 2无需派发 3没有派发
	mysql.DB.Table("transaction_record_models").Select("sum(back_money) as no_back_money").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Where("distribute=3").Scan(&result)
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Where("distribute=3").Count(&NoBackNum)
	//Profit       float64 `gorm:"type:decimal(10,2)"` //盈利
	//mysql.DB.Table("transaction_record_models").Select("sum(result) as profit").Where("symbol=?", JbLx).Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Scan(&result)
	//下注金额
	mysql.DB.Table("transaction_record_models").Select("sum(money) as bet_money").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Scan(&result)

	add := model.DailyData{
		ToAddress:    player.HandAccountAddress,
		BetNum:       BetNum,
		BetPeopleNum: BetPeopleNum,
		WinNum:       WinNum,
		FailNum:      FailNum,
		WinAccount:   result.WinAccount,
		BackMoney:    result.BackMoney,
		NoBackMoney:  result.NoBackMoney,
		BetMoney:     result.BetMoney,
		NoBackNum:    NoBackNum,
		Profit:       result.BetMoney - result.BackMoney,
		Symbol:       "USDT",
		Updated:      time.Now().Unix(),
		Date:         timeDate,
	}
	//判断是否存在今日数据
	daily := model.DailyData{}
	err = mysql.DB.Model(&model.DailyData{}).Where("date=?", timeDate).Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", add.ToAddress).First(&daily).Error
	if err != nil { //没有找到 直接 插入
		add.Created = time.Now().Unix()
		mysql.DB.Save(&add)
	} else {
		//更新
		err = mysql.DB.Model(&model.DailyData{}).Where("id=?", daily.ID).Update(&add).Error
	}

	/**
	xRT
	*/
	var result2 WinAccount
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Count(&BetNum)
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Group("form").Count(&BetPeopleNum)
	////状态 1输  2 赢 3无效
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Count(&WinNum)
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=1").Count(&FailNum)
	//中奖金额   派奖金额
	mysql.DB.Table("transaction_record_models").Select("sum(result) as win_account ,sum(back_money)as back_money").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Scan(&result2)
	//未派奖金额数 是否派发   1 已经派发 2无需派发 3没有派发
	mysql.DB.Table("transaction_record_models").Select("sum(back_money) as no_back_money").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Where("distribute=3").Scan(&result2)
	mysql.DB.Table("transaction_record_models").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Where("status=2").Where("distribute=3").Count(&NoBackNum)
	//Profit       float64 `gorm:"type:decimal(10,2)"` //盈利
	mysql.DB.Table("transaction_record_models").Select("sum(result) as profit").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Scan(&result2)
	//下注金额
	mysql.DB.Table("transaction_record_models").Select("sum(money) as bet_money").Where("(symbol=? or symbol=?)", "TRX", "trx").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Scan(&result2)

	add = model.DailyData{
		ToAddress:    player.HandAccountAddress,
		BetNum:       BetNum,
		BetPeopleNum: BetPeopleNum,
		WinNum:       WinNum,
		FailNum:      FailNum,
		WinAccount:   result2.WinAccount,
		BackMoney:    result2.BackMoney,
		NoBackMoney:  result2.NoBackMoney,
		BetMoney:     result2.BetMoney,
		NoBackNum:    NoBackNum,
		Profit:       result2.BetMoney - result2.BackMoney,
		Symbol:       "TRX",
		Updated:      time.Now().Unix(),
		Date:         timeDate,
	}
	//判断是否存在今日数据
	daily = model.DailyData{}
	err = mysql.DB.Model(&model.DailyData{}).Where("date=?", timeDate).Where("to_address=?", add.ToAddress).Where("symbol=?", "TRX").First(&daily).Error
	if err != nil { //没有找到 直接 插入
		add.Created = time.Now().Unix()
		mysql.DB.Save(&add)
	} else {
		//更新
		mysql.DB.Model(&model.DailyData{}).Where("id=?", daily.ID).Update(&add)
	}

	util.JsonWrite(c, 200, nil, "执行成功")
	return

}

func GetEverydayStatistics(c *gin.Context) {
	action := c.Query("action")
	if action == "GET" {
		//获取 玩法管理
		page, _ := strconv.Atoi(c.Query("page"))
		limit, _ := strconv.Atoi(c.Query("limit"))

		handAccountAddress := c.Query("hand_account_address")

		var total int = 0
		Db := mysql.DB.Where("to_address=?", handAccountAddress)

		if symbol, isExist := c.GetQuery("Symbol"); isExist == true {
			Db = Db.Where("symbol=?", symbol)
		}

		vipEarnings := make([]model.DailyData, 0)
		if status, isExist := c.GetQuery("status"); isExist == true {
			status, _ := strconv.Atoi(status)
			Db = Db.Where("status=?", status)
		}
		Db.Table("daily_data").Count(&total)
		Db = Db.Model(&vipEarnings).Offset((page - 1) * limit).Limit(limit).Order("updated desc")
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
}
