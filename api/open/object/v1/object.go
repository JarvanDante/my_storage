package v1

import "github.com/gogf/gf/v2/frame/g"

type ObjectItem struct {
	Id          string `json:"id"`
	Biz         string `json:"biz"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	PublicUrl   string `json:"public_url"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type ListReq struct {
	g.Meta  `path:"/open/objects" method:"get" tags:"Open/Object" summary:"本站对象列表"`
	Page    int    `json:"page" d:"1"`
	Size    int    `json:"size" d:"20"`
	Keyword string `json:"keyword"`
	Biz     string `json:"biz"`
	Status  int    `json:"status" d:"-1"`
}

type ListRes struct {
	List  []ObjectItem `json:"list"`
	Total int          `json:"total"`
}

type CreateReq struct {
	g.Meta      `path:"/open/objects" method:"post" tags:"Open/Object" summary:"创建对象并返回预签名上传地址"`
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
	g.Meta `path:"/open/objects/{id}" method:"get" tags:"Open/Object" summary:"对象详情"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type DetailRes struct {
	ObjectItem
	ObjectKey string `json:"object_key"`
}

type ConfirmReq struct {
	g.Meta `path:"/open/objects/{id}/confirm" method:"post" tags:"Open/Object" summary:"确认上传完成"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type ConfirmRes struct {
	Id        string `json:"id"`
	Status    int    `json:"status"`
	SizeBytes int64  `json:"size_bytes"`
	PublicUrl string `json:"public_url"`
}

type DownloadURLReq struct {
	g.Meta `path:"/open/objects/{id}/download-url" method:"post" tags:"Open/Object" summary:"预签名下载地址"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type DownloadURLRes struct {
	DownloadUrl string `json:"download_url"`
	Method      string `json:"method"`
	ExpireSec   int    `json:"expire_sec"`
	PublicUrl   string `json:"public_url"`
}

type DeleteReq struct {
	g.Meta `path:"/open/objects/{id}" method:"delete" tags:"Open/Object" summary:"删除本站对象"`
	Id     string `json:"id" in:"path" v:"required|length:16,16"`
}

type DeleteRes struct {
	Id string `json:"id"`
}

type UsageReq struct {
	g.Meta `path:"/open/usage" method:"get" tags:"Open/Object" summary:"本站存储用量"`
}

type UsageRes struct {
	AppKey    string `json:"app_key"`
	MaxBytes  int64  `json:"max_bytes"`
	UsedBytes int64  `json:"used_bytes"`
	MaxFiles  int64  `json:"max_files"`
	UsedFiles int64  `json:"used_files"`
}
