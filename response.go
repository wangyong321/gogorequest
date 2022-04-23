package gogrequest

import (
	"net/http"
	"time"
)

// 同步引擎响应体
type SyncEngineResponse struct {
	Status      bool
	Error       error
	Request     *syncEngineRequestBody
	Response    *http.Response
	StatusCode  int
	Text        string
	StartTime   time.Time
	EndTime     time.Time
	ConsumeTime float64
}

// 异步引擎响应体
type AsyncEngineResponse struct {
	Status      bool
	Error       error
	Request     *asyncEngineRequestBody
	Response    *http.Response
	StatusCode  int
	Text        string
	StartTime   time.Time
	EndTime     time.Time
	ConsumeTime float64
}

// 批量异步响应体
type BatchAsyncEngineResponse struct {
	Status      bool
	Error       error
	Request     *batchAsyncEngineRequestBody
	Response    *http.Response
	StatusCode  int
	Text        string
	StartTime   time.Time
	EndTime     time.Time
	ConsumeTime float64
}
