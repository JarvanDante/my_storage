package controller

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "github.com/JarvanDante/my_storage/api/open/object/v1"
	"github.com/JarvanDante/my_storage/internal/modules/object/domain"
	"github.com/JarvanDante/my_storage/internal/modules/object/service"
	"github.com/JarvanDante/my_storage/internal/shared/errcode"
)

type Open struct {
	svc service.Object
}

func NewOpen(svc service.Object) *Open { return &Open{svc: svc} }

func (c *Open) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	appKey := g.RequestFromCtx(ctx).GetCtxVar("app_key").String()
	list, total, err := c.svc.List(ctx, domain.ListFilter{
		Page: req.Page, Size: req.Size, Keyword: req.Keyword,
		AppKey: appKey, Biz: req.Biz, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.ListRes{Total: total, List: make([]v1.ObjectItem, 0, len(list))}
	for _, o := range list {
		res.List = append(res.List, toOpenItem(o))
	}
	return res, nil
}

func (c *Open) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	r := g.RequestFromCtx(ctx)
	out, err := c.svc.Create(ctx, service.CreateInput{
		AppKey: r.GetCtxVar("app_key").String(), SiteCode: r.GetCtxVar("site_code").String(),
		Biz: req.Biz, Filename: req.Filename, ContentType: req.ContentType,
		SizeBytes: req.SizeBytes, Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{
		Id: out.Id, UploadUrl: out.UploadUrl, Method: out.Method, Bucket: out.Bucket,
		Key: out.Key, ExpireSec: out.ExpireSec, PublicUrl: out.PublicUrl, ContentType: out.ContentType,
	}, nil
}

func (c *Open) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	o, err := c.owned(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DetailRes{ObjectItem: toOpenItem(*o), ObjectKey: o.ObjectKey}, nil
}

func (c *Open) Confirm(ctx context.Context, req *v1.ConfirmReq) (res *v1.ConfirmRes, err error) {
	if _, err := c.owned(ctx, req.Id); err != nil {
		return nil, err
	}
	out, err := c.svc.Confirm(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.ConfirmRes{
		Id: out.Id, Status: out.Status, SizeBytes: out.SizeBytes, PublicUrl: out.PublicUrl,
	}, nil
}

func (c *Open) DownloadURL(ctx context.Context, req *v1.DownloadURLReq) (res *v1.DownloadURLRes, err error) {
	if _, err := c.owned(ctx, req.Id); err != nil {
		return nil, err
	}
	out, err := c.svc.DownloadURL(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DownloadURLRes{
		DownloadUrl: out.DownloadUrl, Method: out.Method, ExpireSec: out.ExpireSec, PublicUrl: out.PublicUrl,
	}, nil
}

func (c *Open) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if _, err := c.owned(ctx, req.Id); err != nil {
		return nil, err
	}
	if err := c.svc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Id: req.Id}, nil
}

func (c *Open) Usage(ctx context.Context, req *v1.UsageReq) (res *v1.UsageRes, err error) {
	appKey := g.RequestFromCtx(ctx).GetCtxVar("app_key").String()
	q, err := c.svc.Quota(ctx, appKey)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return &v1.UsageRes{AppKey: appKey}, nil
	}
	return &v1.UsageRes{
		AppKey: q.AppKey, MaxBytes: q.MaxBytes, UsedBytes: q.UsedBytes,
		MaxFiles: q.MaxFiles, UsedFiles: q.UsedFiles,
	}, nil
}

func (c *Open) owned(ctx context.Context, code string) (*domain.Object, error) {
	appKey := g.RequestFromCtx(ctx).GetCtxVar("app_key").String()
	o, err := c.svc.Get(ctx, code)
	if err != nil {
		return nil, err
	}
	if o == nil || o.AppKey != appKey {
		return nil, gerror.NewCode(errcode.CodeNotFound, "对象不存在")
	}
	return o, nil
}

func toOpenItem(o domain.Object) v1.ObjectItem {
	return v1.ObjectItem{
		Id: o.Code, Biz: o.Biz, Filename: o.Filename, ContentType: o.ContentType,
		SizeBytes: o.SizeBytes, PublicUrl: o.PublicURL, Status: o.Status, CreatedAt: o.CreatedAt,
	}
}
