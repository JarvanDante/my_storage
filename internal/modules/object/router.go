package object

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/JarvanDante/my_storage/internal/modules/object/controller"
	"github.com/JarvanDante/my_storage/internal/modules/object/service"
	openc "github.com/JarvanDante/my_storage/internal/modules/open/controller"
)

func RegisterAdmin(group *ghttp.RouterGroup, svc service.Object) {
	ctrl := controller.NewAdmin(svc)
	group.Bind(ctrl.List, ctrl.Create, ctrl.Upload, ctrl.Detail, ctrl.Confirm, ctrl.DownloadURL, ctrl.Delete, ctrl.Quota)
}

func RegisterOpen(group *ghttp.RouterGroup, svc service.Object) {
	ctrl := openc.NewOpen(svc)
	group.Bind(ctrl.List, ctrl.Create, ctrl.Detail, ctrl.Confirm, ctrl.DownloadURL, ctrl.Delete, ctrl.Usage)
}
