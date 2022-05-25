package gogorequest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// 同步引擎
type SyncEngine struct {
	mainEngine // 继承主引擎
}

func (this *SyncEngine) Visit(method string, targetUrl string, headers map[string]string, body interface{}, timeout time.Duration, proxies string, meta map[string]interface{}) *SyncEngineResponse {
	request := syncEngineRequestBody{
		URL:         targetUrl,
		Method:      method,
		Headers:     headers,
		Body:        body,
		Spider:      this,
		Proxy:       proxies,
		Timeout:     timeout,
		Meta:        meta,
		RetryNumber: 0,
		startTime:   time.Now(),
	}
	return this.get(&request)
}

func (this *SyncEngine) retryVisit(method string, targetUrl string, headers map[string]string, body interface{}, timeout time.Duration, proxies string, meta map[string]interface{}, retryNumber int64, startTime time.Time) *SyncEngineResponse {
	request := syncEngineRequestBody{
		URL:         targetUrl,
		Method:      method,
		Headers:     headers,
		Body:        body,
		Spider:      this,
		Proxy:       proxies,
		Timeout:     timeout,
		Meta:        meta,
		RetryNumber: retryNumber + 1,
		startTime:   startTime,
	}
	return this.get(&request)
}

func (this *SyncEngine) get(request *syncEngineRequestBody) *SyncEngineResponse {
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
			bodyJson, marshalErr := json.Marshal(request.Body)
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
	// 处理返回数据
	return this.onResponse(res, request, request.startTime, endTime, consumeTime)
}

func (this *SyncEngine) onError(res *http.Response, err error, request *syncEngineRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) *SyncEngineResponse {
	var response SyncEngineResponse
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

func (this *SyncEngine) onResponse(res *http.Response, request *syncEngineRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) *SyncEngineResponse {
	var response SyncEngineResponse
	// 读取响应内容
	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return this.onError(res, err, request, startTime, endTime, consumeTime)
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

// 实例化同步引擎
func NewSyncEngine() *SyncEngine {
	s := SyncEngine{}
	s.initTransport()
	return &s
}
