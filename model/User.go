/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package model

type User struct {
	ID                 uint   `gorm:"primaryKey;comment:'主键'"`
	Trc20Address       string //地址
	InvitationCode     string //邀请码  随机生成
	SuperiorId         int    `gorm:"int(10;default:0"` //上级id
	NextSuperiorId     int    `gorm:"int(10;default:0"` //上上级
	NextNextSuperiorId int    `gorm:"int(10;default:0"` //次上上级
	Created            int64  // 注册时间
}




