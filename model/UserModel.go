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
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	eeor "github.com/wangyi/MgHash/error"
	"github.com/wangyi/MgHash/tools"
	"strconv"
	"time"
)

type UserModel struct {
	ID                       uint   `gorm:"primaryKey;comment:'主键'"`
	Trc20Address             string //地址
	InvitationCode           string //邀请码  随机生成
	SuperiorId               int    `gorm:"int(10;default:0"` //上级id
	NextSuperiorId           int    `gorm:"int(10;default:0"` //上上级
	NextNextSuperiorId       int    `gorm:"int(10;default:0"` //次上上级
	TelegramId               int64  `gorm:"default:0"`        //飞机的id
	Nickname                 string //昵称
	Created                  int64  // 注册时间
	TelegramUserName         string
	TheWinningNotification   int `gorm:"default:1"` //1 开启  2关闭   默认开启    (投注通知)
	AdminJurisdiction        int `gorm:"default:1"` //1普通人   2超级管理员!
	UserName                 string
	Password                 string
	Remark                   string   //备注
	VipTotal                 VipTotal `gorm:"-"`
	AllBetMoneyUSDT          float64  `gorm:"-"  `
	AllBackMoneyUSDT         float64  `gorm:"-"  ` //总的返还 金额
	AllBetMoneyTRX           float64  `gorm:"-"  `
	AllBackMoneyTRX          float64  `gorm:"-"  ` //总的返还 金额
	SubordinateBetMoneyUSDT  float64  `gorm:"-"`   //下级的总投注金额
	SubordinateBackMoneyUSDT float64  `gorm:"-"`   //下级的总返还金额
	SubordinateBetMoneyTRX   float64  `gorm:"-"`   //下级的总投注金额
	SubordinateBackMoneyTRx  float64  `gorm:"-"`   //下级的总返还金额

}

type VipTotal struct {
	All   []string
	Today []string
}

func CheckIsExistModelUserModel(db *gorm.DB) {
	if db.HasTable(&UserModel{}) {
		fmt.Println("数据库已经存在了!")
		db.AutoMigrate(&UserModel{})
	} else {
		fmt.Println("数据不存在,所以我要先创建数据库")
		err := db.CreateTable(&UserModel{}).Error
		if err == nil {
			fmt.Println("数据库已经存在了!")
		}
	}
}

//新增用户
func (u *UserModel) AddUser(db *gorm.DB, redis *redis.Client, lang string) (bool, error) {
	//判断用户是否注册过了.
	err := db.Where("trc20_address=?", u.Trc20Address).First(&UserModel{}).Error
	if err == nil {
		if lang == "zh_CN" {
			return false, eeor.OtherError("不要重复注册")
		} else {
			return false, eeor.OtherError("Don't register twice")
		}
	}
	//没有注册就
	u.Created = time.Now().Unix()
	u.InvitationCode = tools.CheckRandStringRunesIsRepetition(redis)
	if u.Trc20Address == "" {
		if lang == "zh_CN" {
			return false, eeor.OtherError("系统错误,注册失败,请重新注册")
		} else {
			return false, eeor.OtherError("Sorry system error, please register again")
		}
	}
	err = db.Save(&u).Error
	if err != nil {
		if lang == "zh_CN" {
			return false, eeor.OtherError("系统错误,注册失败,请重新注册")
		} else {
			return false, eeor.OtherError("Sorry system error, please register again")
		}
	}

	redis.HSet("CheckRandStringRunesIsRepetition", u.InvitationCode, u.Trc20Address)
	return true, nil
}

//返回用户
func (u *UserModel) ReturnUser(db *gorm.DB) (UserModel, error) {

	err := db.Where("id=?", u.ID).First(&u).Error
	if err != nil {
		return *u, err
	}

	return *u, nil
}

//飞机回复 消息内容
func (u *UserModel) ReplyTelegramStart(db *gorm.DB, redis *redis.Client) string {
	err := db.Where("telegram_id=?", u.TelegramId).First(&u).Error
	if err != nil {
		//这个  飞机id 不存在
		redis.HSet("MassTexting", strconv.FormatInt(u.TelegramId, 10), "1")
		redis.HSet("TelegramId_"+strconv.Itoa(int(u.TelegramId)), "status", 1) //1 输入昵称
		return "请输入昵称"
	}

	//昵称为空
	if u.Nickname == "" {
		redis.HSet("TelegramId_"+strconv.Itoa(int(u.TelegramId)), "status", 1) //1 输入昵称
		return "请输入昵称"
	}

	//飞机id 存在  并且有地址  昵称都存在
	if u.TelegramId != 0 && u.Trc20Address != "" && u.Nickname != "" {
		return "请选择快捷操作!"
	}

	//昵称不为空   说明昵称存在
	if u.Trc20Address == "" {
		redis.HSet("TelegramId_"+strconv.Itoa(int(u.TelegramId)), "status", 2) //1 输入昵称
		return "请输入地址"
	}

	return ""
}

//获取邀请码
func (u *UserModel) ReturnInCode(db *gorm.DB) (string, error) {

	user := UserModel{}
	err := db.Where("telegram_id=?", u.TelegramId).First(&user).Error
	if err != nil {
		return "", err
	}

	return user.InvitationCode, nil

}

// start  录入上级
func (u *UserModel) EnteringTelegramId(db *gorm.DB, redis *redis.Client, code string) {
	user := UserModel{}
	err := db.Where("telegram_id=?", u.TelegramId).First(&user).Error
	if err == nil {
		//说明这个小飞机已经存在了
		fmt.Println("说明这个小飞机已经存在了")
		return
	}
	//不存下
	us := UserModel{}
	err = db.Where("invitation_code=?", code).First(&us).Error
	if err != nil {
		//不存在这个邀请码
		fmt.Println("不存在这个邀请码")
		return
	}
	add := UserModel{
		TelegramId:         u.TelegramId,
		Created:            time.Now().Unix(),
		SuperiorId:         int(us.ID),
		NextSuperiorId:     us.SuperiorId,
		NextNextSuperiorId: us.NextSuperiorId,
		TelegramUserName:   u.TelegramUserName,
	}
	err = db.Save(&add).Error
	if err != nil {
		fmt.Println(err.Error())
	}

}

//返回地址

//获取地址
func (u *UserModel) ReturnTrxAddress(db *gorm.DB) (string, error) {
	user := UserModel{}
	err := db.Where("telegram_id=?", u.TelegramId).First(&user).Error
	if err != nil {
		return "", err
	}
	return user.Trc20Address, nil

}

//返回飞机id
func (u *UserModel) ReturnTelegramId(db *gorm.DB) (int64, error) {

	err := db.Where("Trc20_address=?", u.Trc20Address).First(&u).Error
	if err != nil {
		return 0, err
	}

	return u.TelegramId, nil

}

//是否 是超级管理员!
func (u *UserModel) IfAdmin(db *gorm.DB) bool {
	err := db.Where("telegram_id=?", u.TelegramId).Where("user_name=?", u.UserName).Where("password=? and  admin_jurisdiction=?", u.Password, 2).First(&u).Error
	if err != nil {
		return false
	}
	return true
}

//通过 给你 id 判断是否是管理员

func (u *UserModel) IfAdminForTid(db *gorm.DB) bool {
	err := db.Where("telegram_id=?", u.TelegramId).Where("admin_jurisdiction=?", 2).First(&u).Error
	if err != nil {
		return false
	}
	return true
}
