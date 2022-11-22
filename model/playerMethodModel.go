/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type PlayerMethodModel struct {
	ID                 uint    `gorm:"primaryKey;comment:'主键'"`
	Name               string  //玩法名字    牛牛 大小 单双(写法一定要定死)
	Remark             string  //备注
	HandAccountAddress string  //收账地址
	MinBetMoneyForUsdt float64 `gorm:"type:decimal(10,2)"` //最小投注的钱  USDT
	ManBetMoneyForUsdt float64 `gorm:"type:decimal(10,2)"` //最大投注的钱  USDT
	MinBetMoneyForXrt  float64 `gorm:"type:decimal(10,2)"` //最小投注的钱  TRX
	ManBetMoneyForXrt  float64 `gorm:"type:decimal(10,2)"` //最小投注的钱  TRX
	Status             int     //1 启用  2不启用	Updated            int64  //更新时间
	Created            int64   //创建时间
	LossPerCent        float64 `gorm:"type:decimal(10,2)"` //赔率
	Kinds              int     `gorm:"int(10);default:2"`  //    1 对手 2自己

}

func CheckIsExistModelPlayerMethodModel(db *gorm.DB) {
	if db.HasTable(&PlayerMethodModel{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&PlayerMethodModel{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&PlayerMethodModel{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

type TransactionsJsonToStruct struct {
	Data []struct {
		TransactionID string `json:"transaction_id"`
		TokenInfo     struct {
			Symbol   string `json:"symbol"`
			Address  string `json:"address"`
			Decimals int    `json:"decimals"`
			Name     string `json:"name"`
		} `json:"token_info"`
		BlockTimestamp int64  `json:"block_timestamp"`
		From           string `json:"from"`
		To             string `json:"to"`
		Type           string `json:"type"`
		Value          string `json:"value"`
	} `json:"data"`
	Success bool `json:"success"`
	Meta    struct {
		At          int64  `json:"at"`
		Fingerprint string `json:"fingerprint"`
		Links       struct {
			Next string `json:"next"`
		} `json:"links"`
		PageSize int `json:"page_size"`
	} `json:"meta"`
}




type TransactionsJsonToStructForTrx struct {
	Data []struct {
		Ret []struct {
			ContractRet string `json:"contractRet"`
			Fee int `json:"fee"`
		} `json:"ret"`
		Signature []string `json:"signature"`
		TxID string `json:"txID"`
		NetUsage int `json:"net_usage"`
		RawDataHex string `json:"raw_data_hex"`
		NetFee int `json:"net_fee"`
		EnergyUsage int `json:"energy_usage"`
		BlockNumber int `json:"blockNumber"`
		BlockTimestamp int64 `json:"block_timestamp"`
		EnergyFee int `json:"energy_fee"`
		EnergyUsageTotal int `json:"energy_usage_total"`
		RawData struct {
			Contract []struct {
				Parameter struct {
					Value struct {
						Amount int `json:"amount"`
						OwnerAddress string `json:"owner_address"`
						ToAddress string `json:"to_address"`
					} `json:"value"`
					TypeURL string `json:"type_url"`
				} `json:"parameter"`
				Type string `json:"type"`
			} `json:"contract"`
			RefBlockBytes string `json:"ref_block_bytes"`
			RefBlockHash string `json:"ref_block_hash"`
			Expiration int64 `json:"expiration"`
			Timestamp int64 `json:"timestamp"`
		} `json:"raw_data"`
		InternalTransactions []interface{} `json:"internal_transactions"`
	} `json:"data"`
	Success bool `json:"success"`
	Meta struct {
		At int64 `json:"at"`
		Fingerprint string `json:"fingerprint"`
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
		PageSize int `json:"page_size"`
	} `json:"meta"`
}




//获取自己的赔率
func (p *PlayerMethodModel) GetLossPerCent(db *gorm.DB) float64 {

	pp := PlayerMethodModel{}
	err := db.Where("id=?", p.ID).First(&pp).Error
	if err != nil {
		return 0.95
	}

	return pp.LossPerCent

}
