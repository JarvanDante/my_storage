package logic

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	adminv1 "github.com/JarvanDante/my_storage/api/admin/object/v1"
	"github.com/JarvanDante/my_storage/internal/dao"
	"github.com/JarvanDante/my_storage/internal/modules/object/domain"
	"github.com/JarvanDante/my_storage/internal/modules/object/service"
	"github.com/JarvanDante/my_storage/internal/shared/errcode"
	"github.com/JarvanDante/my_storage/internal/shared/storage"
)

var safeSegment = regexp.MustCompile(`[^a-zA-Z0-9._\-]+`)

type sObject struct {
	repo  domain.Repository
	store *storage.Minio
}

func New(repo domain.Repository, store *storage.Minio) service.Object {
	return &sObject{repo: repo, store: store}
}

func (s *sObject) List(ctx context.Context, f domain.ListFilter) ([]domain.Object, int, error) {
	return s.repo.List(ctx, f)
}

func (s *sObject) Get(ctx context.Context, code string) (*domain.Object, error) {
	return s.repo.GetByCode(ctx, code)
}

func (s *sObject) Create(ctx context.Context, in service.CreateInput) (*service.CreateResult, error) {
	if s.store == nil {
		return nil, gerror.NewCode(errcode.CodeBadRequest, "MinIO 未初始化")
	}
	site := sanitizeSegment(in.SiteCode, "site")
	biz := sanitizeSegment(in.Biz, g.Cfg().MustGet(ctx, "storage.default_biz", "common").String())
	filename := sanitizeFilename(in.Filename)
	if filename == "" {
		return nil, gerror.NewCode(errcode.CodeBadRequest, "filename 无效")
	}
	if in.ContentType == "" {
		in.ContentType = guessContentType(filename)
	}

	if in.AppKey != "" {
		if err := s.ensureAndCheckQuota(ctx, in.AppKey, in.SizeBytes); err != nil {
			return nil, err
		}
	}

	bucket := s.store.Bucket()
	obj, err := s.repo.Create(ctx, domain.CreateInput{
		AppKey: in.AppKey, SiteCode: site, Biz: biz, Filename: filename,
		ContentType: in.ContentType, SizeBytes: in.SizeBytes, Bucket: bucket,
		ObjectKey: site + "/" + biz + "/{code}/" + filename,
		Remark:    in.Remark,
	})
	if err != nil {
		return nil, err
	}
	uploadURL, err := s.store.PresignPut(ctx, bucket, obj.ObjectKey)
	if err != nil {
		return nil, gerror.Wrap(err, "生成预签名失败")
	}
	publicURL := s.store.PublicURL(bucket, obj.ObjectKey)
	return &service.CreateResult{
		Id: obj.Code, UploadUrl: uploadURL, Method: "PUT", Bucket: bucket,
		Key: obj.ObjectKey, ExpireSec: s.store.ExpireSec(), PublicUrl: publicURL,
		ContentType: in.ContentType,
	}, nil
}

func (s *sObject) Confirm(ctx context.Context, code string) (*adminv1.ConfirmRes, error) {
	if s.store == nil {
		return nil, gerror.NewCode(errcode.CodeBadRequest, "MinIO 未初始化")
	}
	obj, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, gerror.NewCode(errcode.CodeNotFound, "对象不存在")
	}
	if obj.Status == dao.StatusReady {
		return &adminv1.ConfirmRes{
			Id: obj.Code, Status: obj.Status, SizeBytes: obj.SizeBytes, PublicUrl: obj.PublicURL,
		}, nil
	}
	size, ctype, err := s.store.Stat(ctx, obj.Bucket, obj.ObjectKey)
	if err != nil {
		return nil, gerror.NewCode(errcode.CodeBadRequest, "对象尚未上传或不存在，请先 PUT 到预签名地址")
	}
	if obj.AppKey != "" {
		if err := s.ensureAndCheckQuota(ctx, obj.AppKey, size); err != nil {
			return nil, err
		}
	}
	publicURL := s.store.PublicURL(obj.Bucket, obj.ObjectKey)
	if ctype == "" {
		ctype = obj.ContentType
	}
	if err := s.repo.MarkReady(ctx, code, size, ctype, publicURL); err != nil {
		return nil, err
	}
	if obj.AppKey != "" {
		_ = s.repo.AddUsage(ctx, obj.AppKey, size, 1)
	}
	return &adminv1.ConfirmRes{
		Id: code, Status: dao.StatusReady, SizeBytes: size, PublicUrl: publicURL,
	}, nil
}

func (s *sObject) DownloadURL(ctx context.Context, code string) (*adminv1.DownloadURLRes, error) {
	if s.store == nil {
		return nil, gerror.NewCode(errcode.CodeBadRequest, "MinIO 未初始化")
	}
	obj, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, gerror.NewCode(errcode.CodeNotFound, "对象不存在")
	}
	if obj.Status != dao.StatusReady {
		return nil, gerror.NewCode(errcode.CodeBadRequest, "对象未就绪")
	}
	url, err := s.store.PresignGet(ctx, obj.Bucket, obj.ObjectKey)
	if err != nil {
		return nil, gerror.Wrap(err, "生成下载地址失败")
	}
	return &adminv1.DownloadURLRes{
		DownloadUrl: url, Method: "GET", ExpireSec: s.store.ExpireSec(), PublicUrl: obj.PublicURL,
	}, nil
}

func (s *sObject) Delete(ctx context.Context, code string) error {
	obj, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	if obj == nil {
		return gerror.NewCode(errcode.CodeNotFound, "对象不存在")
	}
	if s.store == nil {
		return gerror.NewCode(errcode.CodeBadRequest, "MinIO 未初始化，拒绝删除以免残留对象")
	}
	if err := s.store.RemoveObject(ctx, obj.Bucket, obj.ObjectKey); err != nil {
		return gerror.Wrap(err, "清理对象存储失败")
	}
	if err := s.repo.SoftDelete(ctx, code); err != nil {
		return err
	}
	if obj.AppKey != "" && obj.Status == dao.StatusReady {
		_ = s.repo.AddUsage(ctx, obj.AppKey, -obj.SizeBytes, -1)
	}
	return nil
}

func (s *sObject) Quota(ctx context.Context, appKey string) (*domain.Quota, error) {
	if appKey == "" {
		return nil, gerror.NewCode(errcode.CodeBadRequest, "app_key 必填")
	}
	maxB := g.Cfg().MustGet(ctx, "storage.default_quota_bytes", int64(10<<30)).Int64()
	maxF := g.Cfg().MustGet(ctx, "storage.default_quota_files", int64(100000)).Int64()
	if err := s.repo.EnsureQuota(ctx, appKey, maxB, maxF); err != nil {
		return nil, err
	}
	return s.repo.GetQuota(ctx, appKey)
}

func (s *sObject) ensureAndCheckQuota(ctx context.Context, appKey string, addBytes int64) error {
	maxB := g.Cfg().MustGet(ctx, "storage.default_quota_bytes", int64(10<<30)).Int64()
	maxF := g.Cfg().MustGet(ctx, "storage.default_quota_files", int64(100000)).Int64()
	if err := s.repo.EnsureQuota(ctx, appKey, maxB, maxF); err != nil {
		return err
	}
	q, err := s.repo.GetQuota(ctx, appKey)
	if err != nil {
		return err
	}
	if q == nil {
		return nil
	}
	if q.MaxFiles > 0 && q.UsedFiles >= q.MaxFiles {
		return gerror.NewCode(errcode.CodeQuotaExceed, "文件数配额已满")
	}
	if q.MaxBytes > 0 && addBytes > 0 && q.UsedBytes+addBytes > q.MaxBytes {
		return gerror.NewCode(errcode.CodeQuotaExceed, "容量配额不足")
	}
	return nil
}

func sanitizeSegment(s, fallback string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		s = fallback
	}
	s = safeSegment.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-.")
	if s == "" {
		return fallback
	}
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}

func sanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, "/", "")
	name = safeSegment.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	if name == "" {
		return "file.bin"
	}
	if len(name) > 200 {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		if len(base) > 180 {
			base = base[:180]
		}
		name = base + ext
	}
	return name
}

func guessContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
