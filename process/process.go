/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package process

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/wangyi/MgHash/dao/mysql"
	"github.com/wangyi/MgHash/model"
	"github.com/wangyi/MgHash/tools"
	"github.com/wangyi/MgHash/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

/**
  服务器重启之后
   初始化  对自己的 账号进行初始化  捡漏
*/

func MyselfInitialize(Db *gorm.DB, redis *redis.Client) {
	Player := make([]model.PlayerMethodModel, 0)
	err := Db.Where("status=1").Where("kinds=2").Find(&Player).Error //kinds=1是 对手 2是自己人
	if err == nil {                                                  //说明查到数据了
		for _, v := range Player {
			go GetTransactionDate(v, redis, Db)
		}
	}
}

/**
     获取USDT  trc20
 */
func GetTransactionDate(Player model.PlayerMethodModel, redis *redis.Client, Db *gorm.DB) {
	minTimestamp, _ := redis.Get("Myself_" + Player.HandAccountAddress).Result()
	if minTimestamp == "" {
		return //结束
	}
	for true {
		//minTimestamp :="1650341413596"   TBkKts1T8uQQHSTT21roFLS7jtn4yuvs99
		url := "https://api.trongrid.io/v1/accounts/" + Player.HandAccountAddress + "/transactions/trc20?only_to=true&limit=200&min_timestamp=" + minTimestamp
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Accept", "application/json")
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		var data model.TransactionsJsonToStruct
		jsonErr := json.Unmarshal(body, &data)
		if jsonErr == nil { //数据入库
			if len(data.Data) > 0 {
				go func() {
					time.Sleep(120 * time.Second) //等待两分钟在执行
					for _, v := range data.Data {
						if v.Type == "Transfer" {
							//判断数据是否存在
							err := Db.Where("transaction_id=?", v.TransactionID).First(&model.TransactionRecordModel{}).Error
							if err != nil { //错误不为空 说明 没有这条数据存在可以插入
								money, _ := util.ToDecimal(v.Value, v.TokenInfo.Decimals).Float64()
								tm := time.Unix(v.BlockTimestamp/1000, 0)
								date := tm.Format("2006-01-02")
								addDate := &model.TransactionRecordModel{
									Symbol:         v.TokenInfo.Symbol,
									Money:          money,
									Created:        time.Now().Unix(),
									Updated:        time.Now().Unix(),
									BlockTimestamp: v.BlockTimestamp,
									TransactionId:  v.TransactionID,
									Form:           v.From,
									ToAddress:      v.To,
									Status:         4,
									Date:           date,
									BetName:        Player.Name,
									Week:           tools.ReturnTheWeek(),
									Month:          tools.ReturnTheMonth(),
								}
								mysql.DB.Save(&addDate)
							}
							if data.Meta.Links.Next == "" { //结束   采集完这里的数据接 结束
								break
							} else {
								minTimestamp = strconv.FormatInt(data.Meta.At, 10)
							}
						}
					}
				}()
			}
		}
		if Player.Kinds == 2 {
			time.Sleep(4 * time.Second) //延迟一秒
		}
		time.Sleep(1 * time.Second) //延迟一秒
	}
}






/***

定时任务

*/

func TotalEvery() {
	for true {
		timeDate := time.Now().Format("2006-01-02")
		players := make([]model.PlayerMethodModel, 0)
		err := mysql.DB.Find(&players).Error
		if err == nil {
			for _, player := range players {
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
				mysql.DB.Table("transaction_record_models").Select("sum(result) as profit").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Scan(&result)
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

				fmt.Println(result2.BetMoney)

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
			}
		}
		time.Sleep(time.Hour * 3)
	}
}

//执行昨天的!
func YesEvery() {
	timeDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	players := make([]model.PlayerMethodModel, 0)
	err := mysql.DB.Find(&players).Error
	if err == nil {
		for _, player := range players {
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
			mysql.DB.Table("transaction_record_models").Select("sum(result) as profit").Where("(symbol=? or symbol=?)", "USDT", "usdt").Where("to_address=?", player.HandAccountAddress).Where("date=?", timeDate).Scan(&result)
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
		}
	}
	time.Sleep(time.Hour * 3)

}
