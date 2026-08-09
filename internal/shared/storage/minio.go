package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	client     *miniogo.Client // 服务端读写（Docker 内多为 host.docker.internal）
	presign    *miniogo.Client // 仅用于预签名，endpoint 必须与浏览器访问 Host 一致
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

	// Region 固定可跳过 GetBucketLocation；Docker 内预签名 client 指向 127.0.0.1 时不能再去拨号探活
	region := g.Cfg().MustGet(ctx, "minio.region", "us-east-1").String()
	if region == "" {
		region = "us-east-1"
	}

	cli, err := miniogo.New(endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio new: %w", err)
	}

	// 预签名按浏览器实际访问的 Host 计算，禁止签完再改 Host（会 SignatureDoesNotMatch）
	presignEndpoint := endpoint
	presignSSL := useSSL
	if publicBase != "" {
		if u, err := url.Parse(publicBase); err == nil && u.Host != "" {
			presignEndpoint = u.Host
			presignSSL = strings.EqualFold(u.Scheme, "https")
		}
	}
	presignCli := cli
	if presignEndpoint != endpoint || presignSSL != useSSL {
		presignCli, err = miniogo.New(presignEndpoint, &miniogo.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: presignSSL,
			Region: region,
		})
		if err != nil {
			return nil, fmt.Errorf("minio presign client: %w", err)
		}
	}

	m := &Minio{
		client: cli, presign: presignCli, endpoint: endpoint, bucket: bucket,
		useSSL: useSSL, publicBase: publicBase, expire: time.Duration(expireSec) * time.Second,
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

// PresignPut 浏览器直传用。勿带未签名的 Content-Type（前端请用 ArrayBuffer body）。
func (m *Minio) PresignPut(ctx context.Context, bucket, key, contentType string) (string, error) {
	_ = contentType
	if bucket == "" {
		bucket = m.bucket
	}
	if err := m.EnsureBucket(ctx, bucket); err != nil {
		return "", err
	}
	u, err := m.presign.PresignedPutObject(ctx, bucket, key, m.expire)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (m *Minio) PresignGet(ctx context.Context, bucket, key string) (string, error) {
	if bucket == "" {
		bucket = m.bucket
	}
	u, err := m.presign.PresignedGetObject(ctx, bucket, key, m.expire, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
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

func (m *Minio) PutObject(ctx context.Context, bucket, key, contentType string, r io.Reader, size int64) error {
	if bucket == "" {
		bucket = m.bucket
	}
	if err := m.EnsureBucket(ctx, bucket); err != nil {
		return err
	}
	opts := miniogo.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}
	_, err := m.client.PutObject(ctx, bucket, key, r, size, opts)
	return err
}

func (m *Minio) ExpireSec() int {
	return int(m.expire.Seconds())
}
