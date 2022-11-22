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

type DailyData struct {
	ID           uint    `gorm:"primaryKey;comment:'主键'"`
	ToAddress    string  //收款地址
	BetNum       int     //下注笔数
	BetPeopleNum int     //下注人数
	WinNum       int     //中奖笔数
	FailNum      int     //未中奖笔数
	WinAccount   float64 `gorm:"type:decimal(10,2)"` //中奖金额
	BackMoney    float64 `gorm:"type:decimal(10,2)"` //派奖金额
	NoBackNum    int     //未派订单个数
	NoBackMoney  float64 `gorm:"type:decimal(10,2)"` //未拍金额
	Profit       float64 `gorm:"type:decimal(10,2)"` //盈利
	Symbol       string  //币种 USDT  or XRT
	BetMoney     float64 `gorm:"type:decimal(10,2)"` //]下注金额
	Updated      int64
	Created      int64
	Date         string
}

func CheckIsExistModelDailyData(db *gorm.DB) {
	if db.HasTable(&DailyData{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&DailyData{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&DailyData{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}
