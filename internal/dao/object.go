package dao

import (
	"context"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"github.com/JarvanDante/my_storage/internal/modules/object/domain"
	"github.com/JarvanDante/my_storage/internal/shared/kit"
)

const (
	StatusPending = 0
	StatusReady   = 1
	StatusDeleted = 2
)

type objectRepo struct{}

func NewObjectRepo() domain.Repository { return &objectRepo{} }

func (r *objectRepo) List(ctx context.Context, f domain.ListFilter) ([]domain.Object, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 100 {
		f.Size = 20
	}
	m := g.DB().Model("storage_object").Ctx(ctx).Safe().Where("status <> ?", StatusDeleted)
	if f.AppKey != "" {
		m = m.Where("app_key", f.AppKey)
	}
	if f.SiteCode != "" {
		m = m.Where("site_code", f.SiteCode)
	}
	if f.Biz != "" {
		m = m.Where("biz", f.Biz)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		m = m.Where("(filename LIKE ? OR code LIKE ?)", kw, kw)
	}
	if f.Status >= 0 {
		m = m.Where("status", f.Status)
	}
	total, err := m.Clone().Count()
	if err != nil {
		return nil, 0, err
	}
	rows, err := m.OrderDesc("id").Page(f.Page, f.Size).All()
	if err != nil {
		return nil, 0, err
	}
	list := make([]domain.Object, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapObject(row.Map()))
	}
	return list, total, nil
}

func (r *objectRepo) Create(ctx context.Context, in domain.CreateInput) (*domain.Object, error) {
	for i := 0; i < 8; i++ {
		code, err := kit.NewPublicID()
		if err != nil {
			return nil, err
		}
		exist, err := g.DB().Model("storage_object").Ctx(ctx).Where("code", code).One()
		if err != nil {
			return nil, err
		}
		if !exist.IsEmpty() {
			continue
		}
		objectKey := in.ObjectKey
		if objectKey == "" {
			objectKey = in.SiteCode + "/" + in.Biz + "/" + code + "/" + in.Filename
		} else if strings.Contains(objectKey, "{code}") {
			objectKey = strings.ReplaceAll(objectKey, "{code}", code)
		}
		publicURL := in.PublicURL
		_, err = g.DB().Model("storage_object").Ctx(ctx).Data(g.Map{
			"code": code, "app_key": in.AppKey, "site_code": in.SiteCode, "biz": in.Biz,
			"filename": in.Filename, "content_type": in.ContentType, "size_bytes": in.SizeBytes,
			"bucket": in.Bucket, "object_key": objectKey, "public_url": publicURL,
			"status": StatusPending, "remark": in.Remark,
			"created_at": gtime.Now(), "updated_at": gtime.Now(),
		}).Insert()
		if err != nil {
			return nil, err
		}
		return r.GetByCode(ctx, code)
	}
	return nil, errors.New("生成对象短码失败")
}

func (r *objectRepo) GetByCode(ctx context.Context, code string) (*domain.Object, error) {
	row, err := g.DB().Model("storage_object").Ctx(ctx).Where("code", code).Where("status <> ?", StatusDeleted).One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, nil
	}
	o := mapObject(row.Map())
	return &o, nil
}

func (r *objectRepo) MarkReady(ctx context.Context, code string, size int64, contentType, publicURL string) error {
	data := g.Map{
		"status": StatusReady, "updated_at": gtime.Now(),
	}
	if size > 0 {
		data["size_bytes"] = size
	}
	if contentType != "" {
		data["content_type"] = contentType
	}
	if publicURL != "" {
		data["public_url"] = publicURL
	}
	_, err := g.DB().Model("storage_object").Ctx(ctx).Where("code", code).Data(data).Update()
	return err
}

func (r *objectRepo) SoftDelete(ctx context.Context, code string) error {
	_, err := g.DB().Model("storage_object").Ctx(ctx).Where("code", code).Data(g.Map{
		"status": StatusDeleted, "updated_at": gtime.Now(),
	}).Update()
	return err
}

func (r *objectRepo) EnsureQuota(ctx context.Context, appKey string, maxBytes, maxFiles int64) error {
	if appKey == "" {
		return nil
	}
	exist, err := g.DB().Model("storage_quota").Ctx(ctx).Where("app_key", appKey).One()
	if err != nil {
		return err
	}
	if !exist.IsEmpty() {
		return nil
	}
	_, err = g.DB().Model("storage_quota").Ctx(ctx).Data(g.Map{
		"app_key": appKey, "max_bytes": maxBytes, "max_files": maxFiles,
		"used_bytes": 0, "used_files": 0, "updated_at": gtime.Now(),
	}).Insert()
	return err
}

func (r *objectRepo) GetQuota(ctx context.Context, appKey string) (*domain.Quota, error) {
	row, err := g.DB().Model("storage_quota").Ctx(ctx).Where("app_key", appKey).One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, nil
	}
	m := row.Map()
	return &domain.Quota{
		AppKey:   appKey,
		MaxBytes: g.NewVar(m["max_bytes"]).Int64(), UsedBytes: g.NewVar(m["used_bytes"]).Int64(),
		MaxFiles: g.NewVar(m["max_files"]).Int64(), UsedFiles: g.NewVar(m["used_files"]).Int64(),
	}, nil
}

func (r *objectRepo) AddUsage(ctx context.Context, appKey string, bytesDelta, filesDelta int64) error {
	if appKey == "" {
		return nil
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		row, err := tx.Model("storage_quota").Ctx(ctx).Where("app_key", appKey).LockUpdate().One()
		if err != nil {
			return err
		}
		if row.IsEmpty() {
			return nil
		}
		usedB := row["used_bytes"].Int64() + bytesDelta
		usedF := row["used_files"].Int64() + filesDelta
		if usedB < 0 {
			usedB = 0
		}
		if usedF < 0 {
			usedF = 0
		}
		_, err = tx.Model("storage_quota").Ctx(ctx).Where("app_key", appKey).Data(g.Map{
			"used_bytes": usedB, "used_files": usedF, "updated_at": gtime.Now(),
		}).Update()
		return err
	})
}

func mapObject(m g.Map) domain.Object {
	return domain.Object{
		Pk: g.NewVar(m["id"]).Int64(), Code: g.NewVar(m["code"]).String(),
		AppKey: g.NewVar(m["app_key"]).String(), SiteCode: g.NewVar(m["site_code"]).String(),
		Biz: g.NewVar(m["biz"]).String(), Filename: g.NewVar(m["filename"]).String(),
		ContentType: g.NewVar(m["content_type"]).String(), SizeBytes: g.NewVar(m["size_bytes"]).Int64(),
		Bucket: g.NewVar(m["bucket"]).String(), ObjectKey: g.NewVar(m["object_key"]).String(),
		PublicURL: g.NewVar(m["public_url"]).String(), Status: g.NewVar(m["status"]).Int(),
		Remark: g.NewVar(m["remark"]).String(), CreatedAt: g.NewVar(m["created_at"]).String(),
		UpdatedAt: g.NewVar(m["updated_at"]).String(),
	}
}
