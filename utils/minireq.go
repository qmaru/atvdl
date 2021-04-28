package utils

import (
	"github.com/aobeom/minireq"
)

// UserAgent 全局 UA
var UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36 Edg/90.0.818.49"

// Minireq 初始化
var Minireq *minireq.MiniRequest

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
	Minireq = minireq.Requests()
}

// NewHTTP 创建一个HTTP Client
func NewHTTP(proxy string) *minireq.MiniRequest {
	request := minireq.Requests()
	if proxy != "" {
		request.Proxy(proxy)
	}
	return request
}
