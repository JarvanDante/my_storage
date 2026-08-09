package domain

import "context"

type Object struct {
	Pk          int64
	Code        string
	AppKey      string
	SiteCode    string
	Biz         string
	Filename    string
	ContentType string
	SizeBytes   int64
	Bucket      string
	ObjectKey   string
	PublicURL   string
	Status      int
	Remark      string
	CreatedAt   string
	UpdatedAt   string
}

type ListFilter struct {
	Page     int
	Size     int
	Keyword  string
	AppKey   string
	SiteCode string
	Biz      string
	Status   int // -1 all
}

type CreateInput struct {
	AppKey      string
	SiteCode    string
	Biz         string
	Filename    string
	ContentType string
	SizeBytes   int64
	Bucket      string
	ObjectKey   string
	PublicURL   string
	Remark      string
}

type Quota struct {
	AppKey    string
	MaxBytes  int64
	UsedBytes int64
	MaxFiles  int64
	UsedFiles int64
}

type Repository interface {
	List(ctx context.Context, f ListFilter) ([]Object, int, error)
	Create(ctx context.Context, in CreateInput) (*Object, error)
	GetByCode(ctx context.Context, code string) (*Object, error)
	MarkReady(ctx context.Context, code string, size int64, contentType, publicURL string) error
	SoftDelete(ctx context.Context, code string) error
	EnsureQuota(ctx context.Context, appKey string, maxBytes, maxFiles int64) error
	GetQuota(ctx context.Context, appKey string) (*Quota, error)
	AddUsage(ctx context.Context, appKey string, bytesDelta, filesDelta int64) error
}
