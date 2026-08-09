package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	client     *miniogo.Client
	endpoint   string
	bucket     string
	useSSL     bool
	publicBase string
	expire     time.Duration
}

func NewMinio(ctx context.Context) (*Minio, error) {
	endpoint := g.Cfg().MustGet(ctx, "minio.endpoint", "127.0.0.1:9000").String()
	accessKey := g.Cfg().MustGet(ctx, "minio.access_key", "minioadmin").String()
	secretKey := g.Cfg().MustGet(ctx, "minio.secret_key", "minioadmin").String()
	bucket := g.Cfg().MustGet(ctx, "minio.bucket", "my-storage").String()
	useSSL := g.Cfg().MustGet(ctx, "minio.use_ssl", false).Bool()
	publicBase := strings.TrimRight(g.Cfg().MustGet(ctx, "minio.public_base", "").String(), "/")
	expireSec := g.Cfg().MustGet(ctx, "minio.presign_expire_sec", 7200).Int64()
	if expireSec <= 0 {
		expireSec = 7200
	}
	cli, err := miniogo.New(endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio new: %w", err)
	}
	m := &Minio{
		client: cli, endpoint: endpoint, bucket: bucket, useSSL: useSSL,
		publicBase: publicBase, expire: time.Duration(expireSec) * time.Second,
	}
	if err := m.EnsureBucket(ctx, bucket); err != nil {
		g.Log().Warningf(ctx, "minio ensure bucket: %v", err)
	}
	return m, nil
}

func (m *Minio) Bucket() string { return m.bucket }

func (m *Minio) EnsureBucket(ctx context.Context, bucket string) error {
	if bucket == "" {
		bucket = m.bucket
	}
	ok, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return m.client.MakeBucket(ctx, bucket, miniogo.MakeBucketOptions{})
}

func (m *Minio) rewriteURL(raw string) string {
	if m.publicBase == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	b, err := url.Parse(m.publicBase)
	if err != nil || b.Host == "" {
		return raw
	}
	u.Scheme = b.Scheme
	u.Host = b.Host
	return u.String()
}

func (m *Minio) PresignPut(ctx context.Context, bucket, key string) (string, error) {
	if bucket == "" {
		bucket = m.bucket
	}
	if err := m.EnsureBucket(ctx, bucket); err != nil {
		return "", err
	}
	u, err := m.client.PresignedPutObject(ctx, bucket, key, m.expire)
	if err != nil {
		return "", err
	}
	return m.rewriteURL(u.String()), nil
}

func (m *Minio) PresignGet(ctx context.Context, bucket, key string) (string, error) {
	if bucket == "" {
		bucket = m.bucket
	}
	u, err := m.client.PresignedGetObject(ctx, bucket, key, m.expire, nil)
	if err != nil {
		return "", err
	}
	return m.rewriteURL(u.String()), nil
}

func (m *Minio) Stat(ctx context.Context, bucket, key string) (int64, string, error) {
	if bucket == "" {
		bucket = m.bucket
	}
	info, err := m.client.StatObject(ctx, bucket, key, miniogo.StatObjectOptions{})
	if err != nil {
		return 0, "", err
	}
	return info.Size, info.ContentType, nil
}

func (m *Minio) RemoveObject(ctx context.Context, bucket, key string) error {
	if bucket == "" {
		bucket = m.bucket
	}
	key = strings.TrimLeft(key, "/")
	if key == "" {
		return nil
	}
	err := m.client.RemoveObject(ctx, bucket, key, miniogo.RemoveObjectOptions{})
	if err != nil && miniogo.ToErrorResponse(err).Code == "NoSuchKey" {
		return nil
	}
	return err
}

func (m *Minio) PublicURL(bucket, key string) string {
	if bucket == "" {
		bucket = m.bucket
	}
	key = strings.TrimLeft(key, "/")
	base := m.publicBase
	if base == "" {
		scheme := "http"
		if m.useSSL {
			scheme = "https"
		}
		base = fmt.Sprintf("%s://%s", scheme, m.endpoint)
	}
	if strings.HasSuffix(base, "/"+bucket) {
		return base + "/" + key
	}
	return fmt.Sprintf("%s/%s/%s", base, bucket, key)
}

func (m *Minio) ExpireSec() int {
	return int(m.expire.Seconds())
}
