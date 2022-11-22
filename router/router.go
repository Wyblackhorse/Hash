/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/
package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/unrolled/secure"
	"github.com/wangyi/MgHash/controller"
	"github.com/wangyi/MgHash/controller/web_api"
	eeor "github.com/wangyi/MgHash/error"
	"github.com/wangyi/MgHash/logger"
	"log"
	"net/http"
)

func Setup() *gin.Engine {

	r := gin.New()
	r.Use(Cors())
	//r.Use(TlsHandler())
	r.Use(eeor.ErrHandler())
	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery(true))
	r.NoMethod(eeor.HandleNotFound)
	r.NoRoute(eeor.HandleNotFound)
	r.Static("/static", "./static")
	r.GET("/transactions", controller.Transactions)
	r.GET("/GetTransactionInfoById", controller.GetTransactionInfoById)
	r.GET("/SetPlay", controller.SetPlay)
	r.GET("/GetRecord", controller.GetRecord)
	r.POST("/testApi", controller.ReceivePush)
	r.POST("/ReceivePush", controller.ReceivePush)
	r.GET("/SetEverydayStatistics", controller.SetEverydayStatistics)
	// GetEverydayStatistics
	r.GET("/GetEverydayStatistics", controller.GetEverydayStatistics)

	r.GET("/getL")
	//获取 用户
	r.GET("/GetUsers", controller.GetUsers)
	rG := r.Group("/web_api")
	{
		//注册
		rG.GET("/register", web_api.Register)
		rG.GET("/login", web_api.Login)
		//GetConfig
		rG.GET("/getConfig", web_api.GetConfig)
		//EeaMoney
		rG.GET("/eeaMoney", web_api.EeaMoney)
		//GetBetsList
		rG.GET("/getBetsList", web_api.GetBetsList)
		//MyInvite
		rG.GET("/myInvite", web_api.MyInvite)

	}

	hops := viper.GetString("app.https")
	if hops == "1" {
		sslPem := viper.GetString("app.sslPem")
		sslKey := viper.GetString("app.sslKey")
		err := r.RunTLS(fmt.Sprintf(":%d", viper.GetInt("app.port")), sslPem, sslKey)
		fmt.Println(err.Error())
	} else {
		fmt.Println("---")
		_ = r.Run(fmt.Sprintf(":%d", viper.GetInt("app.port")))
	}

	return r
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic info is: %v", err)
			}
		}()

		c.Next()
	}
}

func TlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     "localhost:" + fmt.Sprintf(":%d", viper.GetInt("app.port")),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)
		// If there was an error, do not continue.
		if err != nil {
			return
		}
		c.Next()
	}
}
