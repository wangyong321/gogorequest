package gogorequest

import (
	"encoding/json"
	"fmt"
	humanizee "github.com/dustin/go-humanize"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type writeCounter struct {
	Total uint64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}
func (wc writeCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanizee.Bytes(wc.Total))
}

type FileEngine struct {
	mainEngine // 继承主引擎
}

func (this *FileEngine) Visit(method string, targetUrl string, headers map[string]string, body interface{}, timeout time.Duration, proxies string, filepath string) *FileEngineResponse {
	request := fileEngineRequestBody{
		URL:       targetUrl,
		Method:    method,
		Headers:   headers,
		Body:      body,
		Spider:    this,
		Proxy:     proxies,
		Timeout:   timeout,
		FilePath:  filepath,
		startTime: time.Now(),
	}
	return this.get(&request)
}

func (this *FileEngine) get(request *fileEngineRequestBody) *FileEngineResponse {
	client := http.Client{}
	defer client.CloseIdleConnections()

	// 设置代理和Transport
	addProxyAndTransportErr := this.addProxyAndTransport(&client, request.Proxy, request.Timeout)
	if addProxyAndTransportErr != nil {
		return this.onError(nil, addProxyAndTransportErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
	}

	// 如果body不等于nil, 则生成reader类型body
	var payload *strings.Reader
	if request.Body != nil {
		requestBody, isString := request.Body.(string)
		if !isString {
			bodyJson, marshalErr := json.Marshal(requestBody)
			if marshalErr != nil {
				return this.onError(nil, marshalErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
			}
			payload = strings.NewReader(string(bodyJson))
		} else {
			payload = strings.NewReader(requestBody)
		}
	} else {
		payload = nil
	}

	// 包装请求体
	var req *http.Request
	var newRequestErr error
	if payload == nil {
		req, newRequestErr = http.NewRequest(request.Method, request.URL, nil)
		if newRequestErr != nil {
			return this.onError(nil, newRequestErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
		}
	} else {
		req, newRequestErr = http.NewRequest(request.Method, request.URL, payload)
		if newRequestErr != nil {
			return this.onError(nil, newRequestErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
		}
	}

	// 设置请求头
	for h, hv := range request.Headers {
		req.Header.Add(h, hv)
	}

	// 执行请求
	res, doErr := client.Do(req)
	endTime := time.Now()
	consumeTime := endTime.Sub(request.startTime).Seconds()
	if doErr != nil {
		return this.onError(res, doErr, request, request.startTime, endTime, consumeTime)
	}
	defer res.Body.Close()

	counter := &writeCounter{}
	file, openFileErr := os.OpenFile(request.FilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if openFileErr != nil {
		return this.onError(nil, openFileErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
	}
	defer file.Close()
	_, copyErr := io.Copy(file, io.TeeReader(res.Body, counter))
	if copyErr != nil {
		return this.onError(nil, copyErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
	}

	// 处理返回数据
	return this.onResponse(res, request, request.startTime, endTime, consumeTime)
}

func (this *FileEngine) onError(res *http.Response, err error, request *fileEngineRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) *FileEngineResponse {
	var response FileEngineResponse
	if res != nil {
		response.StatusCode = res.StatusCode
	} else {
		response.StatusCode = 10000
	}
	response.Status = false
	response.Error = err
	response.Request = request
	response.StartTime = startTime
	response.EndTime = endTime
	response.ConsumeTime = consumeTime
	return &response
}

func (this *FileEngine) onResponse(res *http.Response, request *fileEngineRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) *FileEngineResponse {
	var response FileEngineResponse
	response.Status = true
	response.Error = nil
	response.Request = request
	response.StatusCode = res.StatusCode
	response.Text = "OK"
	response.StartTime = startTime
	response.EndTime = endTime
	response.ConsumeTime = consumeTime
	return &response
}

// 实例化文件下载引擎
func NewFileEngine() *FileEngine {
	s := FileEngine{}
	s.initTransport()
	return &s
}
