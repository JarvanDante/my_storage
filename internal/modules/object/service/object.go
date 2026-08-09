package service

import (
	"context"

	adminv1 "github.com/JarvanDante/my_storage/api/admin/object/v1"
	"github.com/JarvanDante/my_storage/internal/modules/object/domain"
)

type CreateInput struct {
	AppKey      string
	SiteCode    string
	Biz         string
	Filename    string
	ContentType string
	SizeBytes   int64
	Remark      string
}

type CreateResult struct {
	Id          string
	UploadUrl   string
	Method      string
	Bucket      string
	Key         string
	ExpireSec   int
	PublicUrl   string
	ContentType string
}

type Object interface {
	List(ctx context.Context, f domain.ListFilter) ([]domain.Object, int, error)
	Get(ctx context.Context, code string) (*domain.Object, error)
	Create(ctx context.Context, in CreateInput) (*CreateResult, error)
	Confirm(ctx context.Context, code string) (*adminv1.ConfirmRes, error)
	DownloadURL(ctx context.Context, code string) (*adminv1.DownloadURLRes, error)
	Delete(ctx context.Context, code string) error
	Quota(ctx context.Context, appKey string) (*domain.Quota, error)
}
