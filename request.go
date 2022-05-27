package gogorequest

import (
	"time"
)

// 同步引擎请求体
type syncEngineRequestBody struct {
	URL         string
	Headers     map[string]string
	Method      string
	Body        interface{}
	Spider      *SyncEngine
	Proxy       string
	Timeout     time.Duration
	Meta        map[string]interface{}
	RetryNumber int64
	startTime   time.Time
}

func (this *syncEngineRequestBody) Retry() *SyncEngineResponse {
	return this.Spider.retryVisit(this.Method, this.URL, this.Headers, this.Body, this.Timeout, this.Proxy, this.Meta, this.RetryNumber, this.startTime)
}

// 异步引擎请求体
type asyncEngineRequestBody struct {
	URL         string
	Headers     map[string]string
	Method      string
	Body        interface{}
	Spider      *AsyncEngine
	Proxy       string
	Timeout     time.Duration
	Meta        map[string]interface{}
	RetryNumber int64
	startTime   time.Time
}

func (this *asyncEngineRequestBody) Retry() {
	this.Spider.retryVisit(this.Method, this.URL, this.Headers, this.Body, this.Timeout, this.Proxy, this.Meta, this.RetryNumber, this.startTime)
}

// 文件下载引擎请求体
type fileEngineRequestBody struct {
	URL       string
	Headers   map[string]string
	Method    string
	Body      interface{}
	Spider    *FileEngine
	Proxy     string
	Timeout   time.Duration
	FilePath  string
	startTime time.Time
}

// 批量异步请求体[引擎自用]
type batchAsyncEngineRequestBody struct {
	URL       string
	Headers   map[string]string
	Method    string
	Body      interface{}
	Spider    *BatchAsyncEngine
	Proxy     string
	Timeout   time.Duration
	Meta      map[string]interface{}
	startTime time.Time
}

// 批量异步请求体[用户设置]
type BatchAsyncEngineRequestBody struct {
	URL     string
	Headers map[string]string
	Method  string
	Body    interface{}
	Proxy   string
	Timeout time.Duration
	Meta    map[string]interface{}
}
