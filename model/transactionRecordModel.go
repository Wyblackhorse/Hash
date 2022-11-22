/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package model

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/wangyi/MgHash/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TransactionRecordModel struct {
	ID                  uint    `gorm:"primaryKey;comment:'主键'"`
	PlayerMethodModelId int     //玩法id
	Symbol              string  //交易类型
	TransactionId       string  //交易的hash 值
	Form                string  // 转账地址
	ToAddress           string  //接收地址
	BlockTimestamp      int64   //区块时间戳
	BlockNum            string  //区块高度 唯一值
	BlockHash           string  //区块hash 值
	Status              int     //状态 1输  2 赢 3无效    4已经获取了交易hash 没有获取区高度id   5 获取了区高度id,没有获取区hash  6和局
	Money               float64 `gorm:"type:decimal(10,2)"` //下注金额
	Updated             int64   //更新时间
	Created             int64   //创建时间
	Result              float64 `gorm:"type:decimal(10,2)"` //输赢结果    //+100  -100
	BackMoney           float64 `gorm:"type:decimal(10,2)"` // 退回多少钱
	PlayerRemark        string  `gorm:"-"`
	BankerResult        string  //庄家结果
	PlayerResult        string  //玩家结果
	Bet                 float64
	BetName             string //投注类型
	Distribute          int    //是否派发   1 已经派发 2无需派发 3没有派发
	Date                string //2011-10-17
	Week                int
	Month               int
}

func CheckIsExistModelTransactionRecordModel(db *gorm.DB) {
	if db.HasTable(&TransactionRecordModel{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&TransactionRecordModel{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&TransactionRecordModel{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

/**

  这个要开启进程
*/

func GetBlockNum(DB *gorm.DB) {

	for true {
		data := make([]TransactionRecordModel, 0)
		err := DB.Where("status=4").Find(&data).Error
		if err == nil { //查到数据了
			//第一步去 获取 BlockNum (通过TransactionId)
			for _, v := range data {
				GetTransactionInfoById(v.TransactionId, DB)
			}
		}
		time.Sleep(1 * time.Second)
	}

}

type GetTransactionInfoByIdParam struct {
	Value string `json:"value"`
}

type ReturnGetTransactionInfoByIdData struct {
	ID              string   `json:"id"`
	Fee             int      `json:"fee"`
	BlockNumber     int      `json:"blockNumber"`
	BlockTimeStamp  int64    `json:"blockTimeStamp"`
	ContractResult  []string `json:"contractResult"`
	ContractAddress string   `json:"contract_address"`
	Receipt         struct {
		EnergyFee        int    `json:"energy_fee"`
		EnergyUsageTotal int    `json:"energy_usage_total"`
		NetFee           int    `json:"net_fee"`
		Result           string `json:"result"`
	} `json:"receipt"`
	Log []struct {
		Address string   `json:"address"`
		Topics  []string `json:"topics"`
		Data    string   `json:"data"`
	} `json:"log"`
}

/**
  通过交易 hash 获取 区块高度
*/
func GetTransactionInfoById(TransactionId string, DB *gorm.DB) {
	url := "https://api.trongrid.io/wallet/gettransactioninfobyid"
	var user GetTransactionInfoByIdParam
	user.Value = TransactionId
	b, _ := json.Marshal(user)
	payload := strings.NewReader(string(b))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var data ReturnGetTransactionInfoByIdData
	jsonErr := json.Unmarshal(body, &data)
	if jsonErr == nil { //json 解析成功
		DB.Model(&TransactionRecordModel{}).Where("transaction_id=?", TransactionId).Update(&TransactionRecordModel{BlockNum: strconv.Itoa(data.BlockNumber), Status: 5})
	}
}

/**
  通过 区高度id  获取 区块hash值
*/
type GetBlockHashParam struct {
	Num int `json:"num"`
}

type ReturnGetBlockHashData struct {
	BlockID     string `json:"blockID"`
	BlockHeader struct {
		RawData struct {
			Number         int    `json:"number"`
			TxTrieRoot     string `json:"txTrieRoot"`
			WitnessAddress string `json:"witness_address"`
			ParentHash     string `json:"parentHash"`
			Version        int    `json:"version"`
			Timestamp      int64  `json:"timestamp"`
		} `json:"raw_data"`
		WitnessSignature string `json:"witness_signature"`
	} `json:"block_header"`
	Transactions []interface{} `json:"transactions"`
}

/**8
  获取 区块 hash
*/
func GetBlockHash(BlockNum string, tr TransactionRecordModel, DB *gorm.DB, LossPerCent float64) {
	url := "https://api.trongrid.io/wallet/getblockbynum"
	var param GetBlockHashParam
	param.Num, _ = strconv.Atoi(BlockNum)
	b, _ := json.Marshal(param)
	payload := strings.NewReader(string(b))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(body))
	var data ReturnGetBlockHashData
	jsonErr := json.Unmarshal(body, &data)
	if jsonErr == nil { //json 解析成功
		//首先查询玩法
		player := PlayerMethodModel{}
		err := DB.Where("hand_account_address=?", tr.ToAddress).First(&player).Error
		if err == nil {
			//判断 投注的 币种 //玩种类 区分
			if tr.Symbol == "USDT" {
				if tr.Money < player.MinBetMoneyForUsdt || tr.Money > player.ManBetMoneyForUsdt { //判断本次压住是否有效 (金钱的最大区间和 最小区间)
					//更新为无效订单
					backMoney := tr.Money
					if player.Name == "百家乐" {
						if tr.Money < player.MinBetMoneyForUsdt {
							backMoney = 0
						} else {
							backMoney = backMoney * 0.999
						}
					}
					DB.Model(&TransactionRecordModel{}).Where("id=?", tr.ID).Update(&TransactionRecordModel{Status: 3, Updated: time.Now().Unix(), BlockHash: data.BlockID, BackMoney: backMoney, Result: 0})
					return
				}
				//获取庄家 和  玩家 下注结果
			} else if tr.Symbol == "RTX" {
				if tr.Money < player.MinBetMoneyForXrt || tr.Money > player.ManBetMoneyForXrt { //判断本次压住是否有效 (金钱的最大区间和 最小区间)
					//更新为无效订单
					backMoney := tr.Money
					if player.Name == "百家乐" {
						if tr.Money < player.MinBetMoneyForUsdt {
							backMoney = 0
						} else {
							backMoney = backMoney * 0.999
						}
					}
					DB.Model(&TransactionRecordModel{}).Where("id=?", tr.ID).Update(&TransactionRecordModel{Status: 3, Updated: time.Now().Unix(), BlockHash: data.BlockID, BackMoney: backMoney, Result: 0})
					return
				}
			}

			if player.Name == "牛牛" {
				if data.BlockID == "" {
					return
				}
				dataArray := util.GetNiuNiuResultForBanker(data.BlockID) // 获取玩家的结果
				if len(dataArray) != 2 {
					fmt.Println("解析失败,dataArray 数组长度错误")
					return
				}
				banker := dataArray[0] //庄家
				player := dataArray[1] //玩家
				backerNiu := banker % 10
				playerNiu := player % 10

				if backerNiu == 0 {
					backerNiu = 10
				}
				if playerNiu == 0 {
					playerNiu = 10
				}

				//对比大小
				backMoney := 0.00 //退回的钱
				result := 0.00    //结果

				status := 1                //玩家输
				bet := tr.Money / 10       //投注金额
				if backerNiu > playerNiu { // 庄家大于玩家   庄家赢
					result = -bet * float64(backerNiu)
				} else if backerNiu < playerNiu { //玩家赢
					result = bet * float64(playerNiu) * LossPerCent
					status = 2
				} else if backerNiu == playerNiu { //点数相等 712d8

					startLen := len(data.BlockID) - 5
					newHash := data.BlockID[startLen:]
					//获取庄家点数
					a := string(newHash[0])
					//获取玩家点数
					b := string(newHash[3])
					c, _ := strconv.Atoi(a) //庄家
					d, _ := strconv.Atoi(b) //玩家

					if c == 0 {
						c = 10
					}
					if d == 0 {
						d = 10
					}

					if c == d {
						if c == 1 || c == 2 {
							result = -bet * float64(backerNiu)
						} else {
							//和局
							result = 0
							status = 6 //和局
						}
					} else if c > d {
						//庄家大
						result = -bet * float64(backerNiu)
					} else if c < d {
						//玩家大
						result = bet * float64(playerNiu) * LossPerCent
						status = 2
					}

				}
				backMoney = tr.Money + result
				DB.Model(&TransactionRecordModel{}).Where("id=?", tr.ID).Update(&TransactionRecordModel{Status: status, Updated: time.Now().Unix(), BlockHash: data.BlockID, Result: result, BackMoney: backMoney, BankerResult: util.NiuNumToString(backerNiu), PlayerResult: util.NiuNumToString(playerNiu), Bet: bet})
				RealTimeInformTheNewBet(tr.Form, data.BlockID, tr.BetName, status, tr.Money, bet, backMoney, tr.BlockTimestamp, tr.Symbol, util.NiuNumToString(playerNiu), util.NiuNumToString(backerNiu), DB)
			}
			//其他玩法
			if player.Name == "尾数" || player.Name == "幸运" {
				status := 1 ////状态 1输  2 赢 3无效    4已经获取了交易hash 没有获取区高度id   5 获取了区高度id,没有获取区hash
				backMoney := 0.00
				result := 0.00
				if data.BlockID == "" {
					return
				}

				if util.WsResult(data.BlockID) { //没有中奖
					result = -tr.Money
				} else {
					//中奖了
					result = tr.Money * LossPerCent
					backMoney = tr.Money * (1 + LossPerCent)
					status = 2
				}

				DB.Model(&TransactionRecordModel{}).Where("id=?", tr.ID).Update(&TransactionRecordModel{Status: status, Updated: time.Now().Unix(), BlockHash: data.BlockID, Result: result, BackMoney: backMoney, Bet: tr.Money})
				RealTimeInformTheNewBet(tr.Form, data.BlockID, tr.BetName, status, tr.Money, tr.Money, backMoney, tr.BlockTimestamp, tr.Symbol, "", "", DB)

			}
			if player.Name == "单双" {
				status := 1 ////状态 1输  2 赢 3无效    4已经获取了交易hash 没有获取区高度id   5 获取了区高度id,没有获取区hash
				backMoney := 0.00
				result := 0.00
				if util.DsResult(data.BlockID, tr.Money) { //true  中奖
					result = tr.Money * LossPerCent
					backMoney = tr.Money * (1 + LossPerCent)
					status = 2
				} else {
					result = -tr.Money
				}
				DB.Model(&TransactionRecordModel{}).Where("id=?", tr.ID).Update(&TransactionRecordModel{Status: status, Updated: time.Now().Unix(), BlockHash: data.BlockID, Result: result, BackMoney: backMoney, Bet: tr.Money})
				RealTimeInformTheNewBet(tr.Form, data.BlockID, tr.BetName, status, tr.Money, tr.Money, backMoney, tr.BlockTimestamp, tr.Symbol, "", "", DB)

			}
			if player.Name == "百家乐" {
				status := 1 ////状态 1输  2 赢 3无效    4已经获取了交易hash 没有获取区高度id   5 获取了区高度id,没有获取区hash
				backMoney := 0.00
				result := 0.00
				if data.BlockID == "" {
					return
				}
				//判断押注是否有效  (类型是否)
				betType := int(tr.Money) % 10
				if betType != 1 || betType != 2 || betType != 3 {
					backMoney := tr.Money * 0.999
					DB.Model(&TransactionRecordModel{}).Where("id=?", tr.ID).Update(&TransactionRecordModel{Status: 3, Updated: time.Now().Unix(), BlockHash: data.BlockID, BackMoney: backMoney, Result: 0})
					return
				}

				Bs := util.BjlResult(data.BlockID, betType)
				if Bs > 0 {

					if Bs == 0.999 {
						status = 6
						backMoney = tr.Money * Bs
					} else {
						status = 2
						result = tr.Money * (Bs - 1)
						backMoney = tr.Money * Bs
					}
				}
				DB.Model(&TransactionRecordModel{}).Where("id=?", tr.ID).Update(&TransactionRecordModel{Status: status, Updated: time.Now().Unix(), BlockHash: data.BlockID, Result: result, BackMoney: backMoney, Bet: tr.Money})
				RealTimeInformTheNewBet(tr.Form, data.BlockID, tr.BetName, status, tr.Money, tr.Money, backMoney, tr.BlockTimestamp, tr.Symbol, "", "", DB)
				//发送飞机通知

			}

		}

	}

}

func GetResult(DB *gorm.DB) {
	for true {
		data := make([]TransactionRecordModel, 0)
		err := DB.Where("status=5").Find(&data).Error
		if err == nil { //查到数据了
			//第一步去 获取 BlockNum (通过TransactionId)
			for _, v := range data {
				p := PlayerMethodModel{ID: v.ID}

				LossPerCent := p.GetLossPerCent(DB)
				GetBlockHash(v.BlockNum, v, DB, LossPerCent)
			}
		}

		time.Sleep(1 * time.Second)
	}

}

//获获取盈利 飞机
func (t *TransactionRecordModel) GetTodayOrYestDayGetMoney(db *gorm.DB) string {
	type ReturnData struct {
		TodayUSDT    float64 `json:"today_usdt"`
		TodayTRX     float64 `json:"today_trx"`
		TodayUSDTRun float64 `json:"today_usdt_run"`
		TodayTRXRun  float64 `json:"today_trx_run"`
		YesUSDT      float64 `json:"yes_usdt"`
		YesTRX       float64 `json:"yes_trx"`
	}
	//查看今日是否投注
	tt := make([]TransactionRecordModel, 0)
	db.Where("form=?", t.Form).Where("date=?", time.Now().Format("2006-01-02")).Find(&tt)
	var ps string
	var re ReturnData
	if len(tt) == 0 {
		//每日今日数据
		ps = "今日尚未投注\n"
	} else {
		//查询今日盈利 的 USDT  和 xrt
		//今日盈利的 usdt
		db.Raw("SELECT  SUM(result) as today_usdt  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=?", "USDT", "usdt", time.Now().Format("2006-01-02"), t.Form).Scan(&re)
		//今日盈利的 trx
		db.Raw("SELECT  SUM(result) as today_trx  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=?", "TRX", "trx", time.Now().Format("2006-01-02"), t.Form).Scan(&re)
		//今日流水 usdt
		db.Raw("SELECT  SUM(money) as today_usdt_run  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=? and status !=?", "USDT", "usdt", time.Now().Format("2006-01-02"), t.Form, 3).Scan(&re)
		//今日流水 trx
		db.Raw("SELECT  SUM(result) as today_trx  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=? and status!=?", "TRX", "trx", time.Now().Format("2006-01-02"), t.Form, 3).Scan(&re)

		ps = "今日流水:" + strconv.FormatFloat(re.TodayUSDTRun, 'f', 0, 64) + "USDT 、" + strconv.FormatFloat(re.TodayTRXRun, 'f', 0, 64) + "TRX \n" +
			"今日盈利:" + strconv.FormatFloat(re.TodayUSDT, 'f', 0, 64) + "USDT 、" + strconv.FormatFloat(re.TodayTRX, 'f', 0, 64) + "TRX \n\n"
	}

	//流水
	//昨日投注
	ttd := make([]TransactionRecordModel, 0)
	db.Where("form=?", t.Form).Where("date=?", time.Now().AddDate(0, 0, -1).Format("2006-01-02")).Find(&ttd)

	if len(ttd) == 0 {
		//每日今日数据
		ps = ps + "昨日没有投注"
	} else {
		//昨日盈利的 usdt
		db.Raw("SELECT  SUM(result) as today_usdt  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=?", "USDT", "usdt", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), t.Form).Scan(&re)
		//昨日盈利的 trx
		db.Raw("SELECT  SUM(result) as today_trx  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=?", "TRX", "trx", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), t.Form).Scan(&re)
		//昨天流水 usdt
		db.Raw("SELECT  SUM(money) as today_usdt_run  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=? and status !=?", "USDT", "usdt", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), t.Form, 3).Scan(&re)
		//昨日流水 trx
		db.Raw("SELECT  SUM(result) as today_trx  FROM transaction_record_models  WHERE (symbol=? or symbol= ?) and  date=? and form=? and status!=?", "TRX", "trx", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), t.Form, 3).Scan(&re)
		ps = ps + "昨日流水:" + strconv.FormatFloat(re.TodayUSDTRun, 'f', 0, 64) + "USDT 、" + strconv.FormatFloat(re.TodayTRXRun, 'f', 0, 64) + "TRX \n" +
			"昨日盈利:" + strconv.FormatFloat(re.TodayUSDT, 'f', 0, 64) + "USDT 、" + strconv.FormatFloat(re.TodayTRX, 'f', 0, 64) + "TRX \n\n"
	}
	return ps
}

//获取走势

func (t *TransactionRecordModel) GetTrend(db *gorm.DB) []string {

	tt := make([]TransactionRecordModel, 0)
	var re []string
	db.Where("form=?", t.Form).Where("bet_name=?", t.BetName).Where("status!= ?", 3).Limit(20).Order("created desc").Find(&tt)
	if len(tt) == 0 {
		return re
	}

	if t.BetName == "牛牛" {
		for _, v := range tt {
			re = append(re, util.NiuReturnImageUrl(v.PlayerResult, 1))
			re = append(re, util.NiuReturnImageUrl(v.BankerResult, 2))
		}
	} else if t.BetName == "单双" {
		for _, v := range tt {
			pally, zJ := util.DsResultBetData(v.BlockHash, v.Money)
			re = append(re, util.DsReturnImageUrl(pally))
			re = append(re, util.DsReturnImageUrl(zJ))
		}
	} else if t.BetName == "尾数" || t.BetName == "幸运" {
		for _, v := range tt {
			boolWs := util.WsResult(v.BlockHash)
			if boolWs {
				re = append(re, "picture/Luckly/win.png")
				re = append(re, "picture/Luckly/lose.png")
			} else {
				re = append(re, "picture/Luckly/lose.png")
				re = append(re, "picture/Luckly/win.png")
			}
		}
	} else if t.BetName == "百家乐" {
		for _, v := range tt {
			pally, zJ := util.BjlBetResultForTelegram(v.BlockHash)
			re = append(re, "picture/Baccarat/player/"+pally+".png")
			re = append(re, "picture/Baccarat/banker/"+zJ+".png")
		}
	}

	return re
}

//头数事实通知
func RealTimeInformTheNewBet(Form string, BlockHash string, lickGame string, status int, money float64, bet float64, BackMoney float64, Timestamp int64, Token string, WagerType string, LotteryType string, db *gorm.DB) {
	u := UserModel{Trc20Address: Form}
	Tid, err := u.ReturnTelegramId(db)
	if err == nil {
		var strBool bool
		var filePath string
		var kinds string
		filePath = "real/" + strconv.FormatInt(Tid, 10) + "/" + time.Now().Format("20060112")
		kinds = BlockHash + ".png"
		if lickGame == "牛牛" {
			strBool = util.RealTimeSendResult("picture/Niu/back.jpg", util.NiuReturnImageUrl(WagerType, 1), util.NiuReturnImageUrl(LotteryType, 2), filePath, kinds)

		} else if lickGame == "幸运" {
			if status == 1 {
				//输
				strBool = util.RealTimeSendResult("picture/Luckly/back.jpg", "picture/Luckly/lose.png", "picture/Luckly/win.png", filePath, kinds)
			} else if status == 2 {
				//赢
				strBool = util.RealTimeSendResult("picture/Luckly/back.jpg", "picture/Luckly/win.png", "picture/Luckly/win.png", filePath, kinds)
			}
		} else if lickGame == "单双" {
			pally, zJ := util.DsResultBetData(BlockHash, money)
			strBool = util.RealTimeSendResult("picture/singleOrDouble/back.jpg", util.DsReturnImageUrl(pally), util.DsReturnImageUrl(zJ), filePath, kinds)

		} else if lickGame == "百家乐" {
			pally, zJ := util.BjlBetResultForTelegram(BlockHash)
			strBool = util.RealTimeSendResult("picture/Baccarat/back.jpg", "picture/Baccarat/player/"+pally+".png", "picture/Baccarat/banker/"+zJ+".png", filePath, kinds)

		}

		if strBool == true {
			if status != 3 {
				betString := strconv.FormatFloat(bet, 'f', 0, 64)
				backString := strconv.FormatFloat(BackMoney, 'f', 0, 64)
				tm := time.Unix(Timestamp/1000, 0)
				util.SendPhotoTo(viper.GetString("hash.HashToken"), Tid, filePath+"/"+kinds, status, lickGame, betString+Token, BlockHash, backString+Token, tm.Format("2006-01-02 15:04:05"))

			}
		}

	}
}
