/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/spf13/cobra"
	"github.com/wangyi/MgHash/common"
	"github.com/wangyi/MgHash/dao/mysql"
	"github.com/wangyi/MgHash/dao/redis"
	"github.com/wangyi/MgHash/logger"
	"github.com/wangyi/MgHash/model"
	"github.com/wangyi/MgHash/plane"
	"github.com/wangyi/MgHash/process"
	"github.com/wangyi/MgHash/router"
	"github.com/wangyi/MgHash/setting"
	"github.com/wangyi/MgHash/tools"
	"github.com/wangyi/MgHash/util"
	"github.com/zh-five/xdaemon"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	port     string
	daemon   bool
	rootPath string
)
var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "启动MgHash服务",
	Example: "go-fly server",
	Run:     run,
}

func init() {
	serverCmd.PersistentFlags().StringVarP(&rootPath, "rootPath", "r", "", "程序根目录")
	serverCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "监听端口号")
	serverCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "是否为守护进程模式")
}

func run(cmd *cobra.Command, args []string) {
	//初始化目录
	initDir()
	//初始化守护进程
	initDaemon()

	if noExist, _ := tools.IsFileNotExist(common.LogDirPath); noExist {
		if err := os.MkdirAll(common.LogDirPath, 0777); err != nil {
			log.Println(err.Error())
		}
	}
	isMainUploadExist, _ := tools.IsFileExist(common.UploadDirPath)
	if !isMainUploadExist {
		os.Mkdir(common.UploadDirPath, os.ModePerm)
	}

	//加载配置
	if err := setting.Init(); err != nil {
		fmt.Println("配置文件初始化事变", err)
		return
	}
	//初始化日志
	if err := logger.Init(); err != nil {
		fmt.Println("日志初始化失败", err)
		return
	}

	defer zap.L().Sync() //缓存日志追加到日志文件中

	//链接数据库
	if err := mysql.Init(); err != nil {
		fmt.Println("mysql 链接失败,", err)
		return
	}
	defer mysql.Close()
	//redis 初始化
	//4.初始化redis连接
	if err := redis.Init(); err != nil {
		fmt.Println("redis文件初始化失败：", err)
		return
	}
	defer redis.Close()
	go TransactionInit()
	go process.MyselfInitialize(mysql.DB, redis.Rdb)
	go model.GetBlockNum(mysql.DB)
	go model.GetResult(mysql.DB)

	//统计时间
	go process.TotalEvery() //定时任务

	//注册机器人

	go process.GoRunning(mysql.DB)
	if err := plane.Init(); err != nil {
		fmt.Println("飞机注册失败", err)
		return
	}

	router.Setup()
}
func TransactionInit() {
	var idArray []int
	for true {
		Player := make([]model.PlayerMethodModel, 0)
		err := mysql.DB.Where("status=1").Where("kinds=?", 1).Find(&Player).Error
		if err != nil {
			fmt.Println("TransactionInit 错误:  " + err.Error())
			return
		}
		for _, v := range Player {
			if tools.InArray(idArray, int(v.ID)) == false {
				idArray = append(idArray, int(v.ID))
				go GetTransactionDate(v) //开启携程
				go GetTransactionDateForTrx(v)
			}
		}
		time.Sleep(10 * time.Second)
	}

}

func GetTransactionDate(Player model.PlayerMethodModel) {
	for true {
		url := "https://api.trongrid.io/v1/accounts/" + Player.HandAccountAddress + "/transactions/trc20?only_to=true&limit=200"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Accept", "application/json")
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		var data model.TransactionsJsonToStruct
		jsonErr := json.Unmarshal(body, &data)
		if jsonErr == nil {
			if len(data.Data) > 0 {
				//获取数据并且入库
				//获取 redis 上次的最后一个
				theLastTransactionId, err := redis.Rdb.Get("GetTransactionDate_" + strconv.Itoa(int(Player.ID))).Result()
				if err != nil {
					//   数据不存在
					//fmt.Println(theLastTransactionId)
					for t, k := range data.Data {
						if k.Type == "Transfer" {
							if t == 0 {
								if theLastTransactionId == k.TransactionID { //没有新的数据更新
									//	fmt.Println("没有新的数据更新")
									break
								} else {
									redis.Rdb.Set("GetTransactionDate_"+strconv.Itoa(int(Player.ID)), k.TransactionID, 0)
								}
							}
							//判断整条数据是否已经存在了不要重复入库
							err = mysql.DB.Where("transaction_id=?", k.TransactionID).First(&model.TransactionRecordModel{}).Error
							if err != nil { //错误不为空 说明 没有这条数据存在可以插入
								money, _ := util.ToDecimal(k.Value, k.TokenInfo.Decimals).Float64()
								addDate := &model.TransactionRecordModel{
									PlayerMethodModelId: int(Player.ID),
									Symbol:              k.TokenInfo.Symbol,
									Money:               money,
									Created:             time.Now().Unix(),
									Updated:             time.Now().Unix(),
									BlockTimestamp:      k.BlockTimestamp,
									TransactionId:       k.TransactionID,
									Form:                k.From,
									ToAddress:           k.To,
									Status:              4,
								}
								mysql.DB.Save(&addDate)
							}
						}
					}
				} else {
					for t, k := range data.Data {
						if k.Type == "Transfer" {
							if t == 0 {
								if theLastTransactionId == k.TransactionID {
									//没有新的数据更新
									//	fmt.Println("没有新的数据更新")
									break
								} else {
									redis.Rdb.Set("GetTransactionDate_"+strconv.Itoa(int(Player.ID)), k.TransactionID, 0)
								}
							}
							//判断整条数据是否已经存在了不要重复入库
							err = mysql.DB.Where("transaction_id=?", k.TransactionID).First(&model.TransactionRecordModel{}).Error
							if err != nil { //错误不为空 说明 没有这条数据存在可以插入
								money, _ := util.ToDecimal(k.Value, k.TokenInfo.Decimals).Float64()
								tm := time.Unix(k.BlockTimestamp/1000, 0)
								date := tm.Format("2006-01-02")
								addDate := &model.TransactionRecordModel{
									PlayerMethodModelId: int(Player.ID),
									Symbol:              k.TokenInfo.Symbol,
									Money:               money,
									Created:             time.Now().Unix(),
									Updated:             time.Now().Unix(),
									BlockTimestamp:      k.BlockTimestamp,
									TransactionId:       k.TransactionID,
									Form:                k.From,
									ToAddress:           k.To,
									Status:              4,
									Date:                date,
								}
								mysql.DB.Save(&addDate)
							}
						}

					}
				}
			}
		}
	}
	time.Sleep(3 * time.Second) //延迟一秒
}

//
func GetTransactionDateForTrx(Player model.PlayerMethodModel) {
	for true {
		url := "https://api.trongrid.io/v1/accounts/" + Player.HandAccountAddress + "/transactions?only_to=true&limit=200"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Accept", "application/json")
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		var data model.TransactionsJsonToStructForTrx
		jsonErr := json.Unmarshal(body, &data)
		if jsonErr == nil {
			if len(data.Data) > 0 {
				//获取数据并且入库
				//获取 redis 上次的最后一个
				theLastTransactionId, err := redis.Rdb.Get("GetTransactionDateForTrx_" + strconv.Itoa(int(Player.ID))).Result()
				if err != nil {
					//   数据不存在
					//fmt.Println(theLastTransactionId)
					for t, k := range data.Data {
						if k.RawData.Contract[0].Type == "TransferContract" || k.RawData.Contract[0].Type == "TransferAssetContract" {
							if t == 0 {
								if theLastTransactionId == k.TxID { //没有新的数据更新
									//	fmt.Println("没有新的数据更新")
									break
								} else {
									redis.Rdb.Set("GetTransactionDateForTrx_"+strconv.Itoa(int(Player.ID)), k.TxID, 0)
								}
							}
							//判断整条数据是否已经存在了不要重复入库
							err = mysql.DB.Where("transaction_id=?", k.TxID).First(&model.TransactionRecordModel{}).Error
							if err != nil { //错误不为空 说明 没有这条数据存在可以插入
								money, _ := util.ToDecimal(strconv.Itoa(k.RawData.Contract[0].Parameter.Value.Amount), 6).Float64()



								addDate := &model.TransactionRecordModel{
									PlayerMethodModelId: int(Player.ID),
									Symbol:              "TRX",
									Money:               money,
									Created:             time.Now().Unix(),
									Updated:             time.Now().Unix(),
									BlockTimestamp:      k.BlockTimestamp,
									TransactionId:       k.TxID,
									Form:                address.HexToAddress(k.RawData.Contract[0].Parameter.Value.OwnerAddress).String(),
									ToAddress:           address.HexToAddress(k.RawData.Contract[0].Parameter.Value.ToAddress).String(),
									Status:              4,
								}
								mysql.DB.Save(&addDate)
							}
						}

					}
				} else {
					for t, k := range data.Data {
						if k.RawData.Contract[0].Type == "TransferContract" || k.RawData.Contract[0].Type == "TransferAssetContract"  {
							if t == 0 {
								if theLastTransactionId == k.TxID {
									//没有新的数据更新
									//	fmt.Println("没有新的数据更新")
									break
								} else {
									redis.Rdb.Set("GetTransactionDateForTrx_"+strconv.Itoa(int(Player.ID)), k.TxID, 0)
								}
							}
							//判断整条数据是否已经存在了不要重复入库
							err = mysql.DB.Where("transaction_id=?", k.TxID).First(&model.TransactionRecordModel{}).Error
							if err != nil { //错误不为空 说明 没有这条数据存在可以插入
								money, _ := util.ToDecimal(strconv.Itoa(k.RawData.Contract[0].Parameter.Value.Amount), 6).Float64()
								tm := time.Unix(k.BlockTimestamp/1000, 0)
								date := tm.Format("2006-01-02")
								addDate := &model.TransactionRecordModel{
									PlayerMethodModelId: int(Player.ID),
									Symbol:              "TRX",
									Money:               money,
									Created:             time.Now().Unix(),
									Updated:             time.Now().Unix(),
									BlockTimestamp:      k.BlockTimestamp,
									TransactionId:       k.TxID,
									Form:                address.HexToAddress(k.RawData.Contract[0].Parameter.Value.OwnerAddress).String(),
									ToAddress:           address.HexToAddress(k.RawData.Contract[0].Parameter.Value.ToAddress).String(),
									Status:              4,
									Date:                date,
								}
								mysql.DB.Save(&addDate)
							}
						}

					}
				}
			}
		}
	}
	time.Sleep(2 * time.Second) //延迟一秒
}

//初始化目录
func initDir() {

	if rootPath == "" {
		rootPath = tools.GetRootPath()
	}
	log.Println("程序运行路径:" + rootPath)
	common.RootPath = rootPath
	common.LogDirPath = rootPath + "/logs/"
	common.ConfigDirPath = rootPath + "/config/"
	common.StaticDirPath = rootPath + "/static/"
	common.UploadDirPath = rootPath + "/static/upload/"

}

//初始化守护进程
func initDaemon() {
	if daemon == true {
		d := xdaemon.NewDaemon(common.LogDirPath + "MgHash.log")
		d.MaxError = 10
		d.Run()
	}
	//记录pid
	ioutil.WriteFile(common.RootPath+"/MgHash.sock", []byte(fmt.Sprintf("%d,%d", os.Getppid(), os.Getpid())), 0666)
}
