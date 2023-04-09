package server

import (
	"context"
	"duwhy/core"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ServerOption struct {
	IProvider core.IProvider

	Server ServerConfig
}

type ServerConfig struct {
	Host string
	Port int
	Auth struct {
		Enable   bool
		UserName string
		Password string
	}
}

func (so *ServerOption) Valid() error {
	if so.IProvider == nil {
		return errors.New("IProvider should not be nil")
	}

	if so.Server.Host == "" {
		so.Server.Host = "127.0.0.1"
	}

	if so.Server.Port == 0 {
		so.Server.Port = 8080
	}

	if so.Server.Auth.Enable {
		if so.Server.Auth.UserName == "" || so.Server.Auth.Password == "" {
			return errors.New("UserName or Password should not be nil when Auth enabled")
		}
	}

	return nil
}

func Serve(ctx context.Context, opt ServerOption) error {
	if err := opt.Valid(); err != nil {
		return err
	}

	server := NewDuHttpServer(&DuServer{opt.IProvider}, opt.Server)

	return server.Run(fmt.Sprintf("%s:%d", opt.Server.Host, opt.Server.Port))
}

func NewDuHttpServer(isrv IDuServer, cfg ServerConfig) *gin.Engine {
	engine := gin.Default()

	if cfg.Auth.Enable {
		engine.Use(gin.BasicAuth(gin.Accounts{cfg.Auth.UserName: cfg.Auth.Password}))
	}

	v1 := engine.Group("api/v1/")
	v1.GET("info", isrv.GetInfoByPath)

	return engine
}

type IDuServer interface {
	GetInfoByPath(c *gin.Context)
}

type GetInfoParams struct {
	PathName string // 路径

	MaxItems        int     // 最大 item 数量
	LongTailPercent float64 // 长尾的切割百分比，例如 95% 意味着排序的后 5% 会被放入 OtherInfo

	Deep int // 深入子节点
}

type DuServer struct {
	provider core.IProvider
}

func (ds *DuServer) GetInfoByPath(c *gin.Context) {
	p := &GetInfoParams{}
	err := c.BindQuery(p)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	i, err := ds.provider.GetInfoByPath(p.PathName, &core.InfoOption{Deep: p.Deep, MaxItems: p.MaxItems, LongTailPercent: p.LongTailPercent})
	if err != nil {
		c.AbortWithError(400, gin.Error{Err: err, Type: gin.ErrorTypeAny})
		return
	}

	c.JSON(200, i)
}
