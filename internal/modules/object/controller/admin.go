package controller

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "github.com/JarvanDante/my_storage/api/admin/object/v1"
	"github.com/JarvanDante/my_storage/internal/modules/object/domain"
	"github.com/JarvanDante/my_storage/internal/modules/object/service"
	"github.com/JarvanDante/my_storage/internal/shared/errcode"
)

type Admin struct {
	svc service.Object
}

func NewAdmin(svc service.Object) *Admin { return &Admin{svc: svc} }

func (c *Admin) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	list, total, err := c.svc.List(ctx, domain.ListFilter{
		Page: req.Page, Size: req.Size, Keyword: req.Keyword,
		AppKey: req.AppKey, SiteCode: req.SiteCode, Biz: req.Biz, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.ListRes{Total: total, List: make([]v1.ObjectItem, 0, len(list))}
	for _, o := range list {
		res.List = append(res.List, toAdminItem(o))
	}
	return res, nil
}

func (c *Admin) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	out, err := c.svc.Create(ctx, service.CreateInput{
		AppKey: req.AppKey, SiteCode: req.SiteCode, Biz: req.Biz,
		Filename: req.Filename, ContentType: req.ContentType,
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

func (c *Admin) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	o, err := c.svc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, gerror.NewCode(errcode.CodeNotFound, "对象不存在")
	}
	return &v1.DetailRes{
		ObjectItem: toAdminItem(*o), Bucket: o.Bucket, ObjectKey: o.ObjectKey,
	}, nil
}

func (c *Admin) Confirm(ctx context.Context, req *v1.ConfirmReq) (res *v1.ConfirmRes, err error) {
	return c.svc.Confirm(ctx, req.Id)
}

func (c *Admin) DownloadURL(ctx context.Context, req *v1.DownloadURLReq) (res *v1.DownloadURLRes, err error) {
	return c.svc.DownloadURL(ctx, req.Id)
}

func (c *Admin) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err := c.svc.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{Id: req.Id}, nil
}

func (c *Admin) Quota(ctx context.Context, req *v1.QuotaReq) (res *v1.QuotaRes, err error) {
	q, err := c.svc.Quota(ctx, req.AppKey)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return &v1.QuotaRes{AppKey: req.AppKey}, nil
	}
	return &v1.QuotaRes{
		AppKey: q.AppKey, MaxBytes: q.MaxBytes, UsedBytes: q.UsedBytes,
		MaxFiles: q.MaxFiles, UsedFiles: q.UsedFiles,
	}, nil
}

func toAdminItem(o domain.Object) v1.ObjectItem {
	return v1.ObjectItem{
		Id: o.Code, AppKey: o.AppKey, SiteCode: o.SiteCode, Biz: o.Biz,
		Filename: o.Filename, ContentType: o.ContentType, SizeBytes: o.SizeBytes,
		PublicUrl: o.PublicURL, Status: o.Status, Remark: o.Remark, CreatedAt: o.CreatedAt,
	}
}
