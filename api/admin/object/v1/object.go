package v1

import "github.com/gogf/gf/v2/frame/g"

type ObjectItem struct {
	Id          string `json:"id"` // 对外短码
	AppKey      string `json:"app_key"`
	SiteCode    string `json:"site_code"`
	Biz         string `json:"biz"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	PublicUrl   string `json:"public_url"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
	CreatedAt   string `json:"created_at"`
}

type ListReq struct {
	g.Meta   `path:"/admin/objects" method:"get" tags:"Admin/Object" summary:"对象列表"`
	Page     int    `json:"page" d:"1"`
	Size     int    `json:"size" d:"20"`
	Keyword  string `json:"keyword"`
	AppKey   string `json:"app_key"`
	SiteCode string `json:"site_code"`
	Biz      string `json:"biz"`
	Status   int    `json:"status" d:"-1"`
}

type ListRes struct {
	List  []ObjectItem `json:"list"`
	Total int          `json:"total"`
}

type CreateReq struct {
	g.Meta      `path:"/admin/objects" method:"post" tags:"Admin/Object" summary:"创建对象并返回预签名上传地址"`
	AppKey      string `json:"app_key"`
	SiteCode    string `json:"site_code" v:"required#site_code必填"`
	Biz         string `json:"biz"`
	Filename    string `json:"filename" v:"required#filename必填"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	Remark      string `json:"remark"`
}

type CreateRes struct {
	Id          string `json:"id"`
	UploadUrl   string `json:"upload_url"`
	Method      string `json:"method"`
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	ExpireSec   int    `json:"expire_sec"`
	PublicUrl   string `json:"public_url"`
	ContentType string `json:"content_type"`
}

type DetailReq struct {
	g.Meta `path:"/admin/objects/{id}" method:"get" tags:"Admin/Object" summary:"对象详情"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type DetailRes struct {
	ObjectItem
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"object_key"`
}

type ConfirmReq struct {
	g.Meta `path:"/admin/objects/{id}/confirm" method:"post" tags:"Admin/Object" summary:"确认上传完成"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type ConfirmRes struct {
	Id        string `json:"id"`
	Status    int    `json:"status"`
	SizeBytes int64  `json:"size_bytes"`
	PublicUrl string `json:"public_url"`
}

type DownloadURLReq struct {
	g.Meta `path:"/admin/objects/{id}/download-url" method:"post" tags:"Admin/Object" summary:"预签名下载地址"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type DownloadURLRes struct {
	DownloadUrl string `json:"download_url"`
	Method      string `json:"method"`
	ExpireSec   int    `json:"expire_sec"`
	PublicUrl   string `json:"public_url"`
}

type DeleteReq struct {
	g.Meta `path:"/admin/objects/{id}" method:"delete" tags:"Admin/Object" summary:"删除对象"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type DeleteRes struct {
	Id string `json:"id"`
}

type QuotaReq struct {
	g.Meta `path:"/admin/quota" method:"get" tags:"Admin/Object" summary:"查询配额用量"`
	AppKey string `json:"app_key" v:"required#app_key必填"`
}

type QuotaRes struct {
	AppKey    string `json:"app_key"`
	MaxBytes  int64  `json:"max_bytes"`
	UsedBytes int64  `json:"used_bytes"`
	MaxFiles  int64  `json:"max_files"`
	UsedFiles int64  `json:"used_files"`
}

// UploadReq 总后台直传（服务端写入 MinIO，避免浏览器预签名签名问题）
type UploadReq struct {
	g.Meta   `path:"/admin/objects/upload" method:"post" mime:"multipart/form-data" tags:"Admin/Object" summary:"服务端代传上传"`
	AppKey   string `json:"app_key" form:"app_key"`
	SiteCode string `json:"site_code" form:"site_code" v:"required#site_code必填"`
	Biz      string `json:"biz" form:"biz"`
	Remark   string `json:"remark" form:"remark"`
}

type UploadRes struct {
	Id        string `json:"id"`
	Status    int    `json:"status"`
	SizeBytes int64  `json:"size_bytes"`
	PublicUrl string `json:"public_url"`
	Filename  string `json:"filename"`
}
