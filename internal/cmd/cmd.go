// Package cmd 启动装配。
package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/JarvanDante/my_storage/internal/dao"
	"github.com/JarvanDante/my_storage/internal/modules/client"
	"github.com/JarvanDante/my_storage/internal/modules/health"
	"github.com/JarvanDante/my_storage/internal/modules/object"
	"github.com/JarvanDante/my_storage/internal/modules/object/logic"
	"github.com/JarvanDante/my_storage/internal/shared/middleware"
	"github.com/JarvanDante/my_storage/internal/shared/storage"
)

// Main 统一存储 API(:8005)。
var Main = gcmd.Command{
	Name:  "storageapi",
	Brief: "统一存储(PaaS) API",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		store, err := storage.NewMinio(ctx)
		if err != nil {
			g.Log().Warningf(ctx, "minio init failed: %v (upload/download 将不可用)", err)
		}
		repo := dao.NewObjectRepo()
		svc := logic.New(repo, store)
		clientRepo := dao.NewClientRepo()

		s := g.Server()
		s.Use(middleware.CORS, ghttp.MiddlewareHandlerResponse)
		s.BindStatusHandler(404, middleware.NotFound)

		health.Register(s.Group("/"))

		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(middleware.AdminToken)
			object.RegisterAdmin(group, svc)
			client.RegisterAdmin(group, clientRepo)
		})

		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(middleware.AppKey)
			object.RegisterOpen(group, svc)
		})

		s.Run()
		return nil
	},
}
