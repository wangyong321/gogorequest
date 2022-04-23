/*
	同步下载引擎
*/
package engine

import (
	"encoding/json"
	req "gogorequest/request.struct"
	rep "gogorequest/response.struct"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type SyncEngine struct {
	mainEngine
}

func (se *SyncEngine) Visit(method string, targetUrl string, headers map[string]string, body interface{}, timeout time.Duration, proxies string, meta map[string]interface{}) *rep.SyncResponse {
	request := req.SyncRequestBody{
		URL:         targetUrl,
		Method:      method,
		Headers:     headers,
		Body:        body,
		Spider:      se,
		Proxy:       proxies,
		Timeout:     timeout,
		Meta:        meta,
		RetryNumber: 0,
		StartTime:   time.Now(),
	}
	return se.get(&request)
}

func (se *SyncEngine) retryVisit(method string, targetUrl string, headers map[string]string, body interface{}, timeout time.Duration, proxies string, meta map[string]interface{}, retryNumber int64, startTime time.Time) *rep.SyncResponse {
	request := req.SyncRequestBody{
		URL:         targetUrl,
		Method:      method,
		Headers:     headers,
		Body:        body,
		Spider:      se,
		Proxy:       proxies,
		Timeout:     timeout,
		Meta:        meta,
		RetryNumber: retryNumber + 1,
		StartTime:   startTime,
	}
	return se.get(&request)
}

func (se *SyncEngine) get(request *req.SyncRequestBody) *rep.SyncResponse {
	client := http.Client{}
	defer client.CloseIdleConnections()

	// 设置代理和Transport
	addProxyAndTransportErr := se.addProxyAndTransport(&client, request.Proxy, request.Timeout)
	if addProxyAndTransportErr != nil {
		return se.onError(nil, addProxyAndTransportErr, request, request.StartTime, time.Now(), time.Now().Sub(request.StartTime).Seconds())
	}

	// 如果body不等于nil, 则生成reader类型body
	var payload *strings.Reader
	if request.Body != nil {
		requestBody, isString := request.Body.(string)
		if !isString {
			bodyJson, marshalErr := json.Marshal(requestBody)
			if marshalErr != nil {
				return se.onError(nil, marshalErr, request, request.StartTime, time.Now(), time.Now().Sub(request.StartTime).Seconds())
			}
			payload = strings.NewReader(string(bodyJson))
		} else {
			payload = strings.NewReader(requestBody)
		}
	} else {
		payload = nil
	}

	// 包装请求体
	var r *http.Request
	var newRequestErr error
	if payload == nil {
		r, newRequestErr = http.NewRequest(request.Method, request.URL, nil)
		if newRequestErr != nil {
			return se.onError(nil, newRequestErr, request, request.StartTime, time.Now(), time.Now().Sub(request.StartTime).Seconds())
		}
	} else {
		r, newRequestErr = http.NewRequest(request.Method, request.URL, payload)
		if newRequestErr != nil {
			return se.onError(nil, newRequestErr, request, request.StartTime, time.Now(), time.Now().Sub(request.StartTime).Seconds())
		}
	}

	// 设置请求头
	for h, hv := range request.Headers {
		r.Header.Add(h, hv)
	}

	// 执行请求
	res, doErr := client.Do(r)
	endTime := time.Now()
	consumeTime := endTime.Sub(request.StartTime).Seconds()
	if doErr != nil {
		return se.onError(res, doErr, request, request.StartTime, endTime, consumeTime)
	}
	defer res.Body.Close()
	// 处理返回数据
	return se.onResponse(res, request, request.StartTime, endTime, consumeTime)
}

func (se *SyncEngine) onError(res *http.Response, err error, request *req.SyncRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) *rep.SyncResponse {
	var response rep.SyncResponse
	if res != nil {
		response.StatusCode = res.StatusCode
	} else {
		response.StatusCode = 10000
	}
	response.Status = false
	response.Error = err
	response.Request = request
	response.Response = res
	response.StartTime = startTime
	response.EndTime = endTime
	response.ConsumeTime = consumeTime
	return &response
}

func (se *SyncEngine) onResponse(res *http.Response, request *req.SyncRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) *rep.SyncResponse {
	var response rep.SyncResponse
	// 读取响应内容
	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return se.onError(res, err, request, startTime, endTime, consumeTime)
	}
	response.Status = true
	response.Error = nil
	response.Request = request
	response.Response = res
	response.StatusCode = res.StatusCode
	response.Text = string(text)
	response.StartTime = startTime
	response.EndTime = endTime
	response.ConsumeTime = consumeTime
	return &response
}

// NewSyncEngine 实例化同步引擎
func NewSyncEngine() *SyncEngine {
	s := SyncEngine{}
	s.initTransport()
	return &s
}
