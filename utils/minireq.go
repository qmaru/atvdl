package utils

import (
	"github.com/qmaru/minireq/v2"
)

// UserAgent 全局 UA
var UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.134 Safari/537.36 Edg/103.0.1264.77"

// Minireq 初始化
var Minireq *minireq.HttpClient

// MiniHeaders Headers
type MiniHeaders = minireq.Headers

// MiniParams Params
type MiniParams = minireq.Params

// MiniJSONData JSONData
type MiniJSONData = minireq.JSONData

// MiniFormData FormData
type MiniFormData = minireq.FormData

// MiniAuth AuthData
type MiniAuth = minireq.Auth

func init() {
	Minireq = minireq.NewClient()
}
