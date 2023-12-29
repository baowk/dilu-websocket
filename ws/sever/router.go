package sever

import (
	"github.com/gin-gonic/gin"

	"github.com/baowk/dilu-core/core"
)

var (
	routerNoCheckRole = make([]func(*gin.RouterGroup), 0)
	//routerCheckRole   = make([]func(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware), 0)
)

// InitRouter 路由初始化
func InitRouter() {
	r := core.GetGinEngine()
	noCheckRoleRouter(r)
}

// noCheckRoleRouter 无需认证的路由
func noCheckRoleRouter(r *gin.Engine) {
	// 可根据业务需求来设置接口版本
	v := r.Group("")

	for _, f := range routerNoCheckRole {
		f(v)
	}
}

func init() {
	routerNoCheckRole = append(routerNoCheckRole, registerPayMerchantRouter)
}

// 默认需登录认证的路由
func registerPayMerchantRouter(v1 *gin.RouterGroup) {
	//v1.StaticFile("/", "resources/index.html")

	v1.GET("ws", wsHandler.ConnectGin)
}
