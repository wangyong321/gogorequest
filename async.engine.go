package gogorequest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// 异步引擎
type AsyncEngine struct {
	mainEngine       // 继承主引擎
	limiter          chan bool
	chanRequest      chan *asyncEngineRequestBody
	chanRetryRequest chan *asyncEngineRequestBody
	ChanResponses    chan *AsyncEngineResponse
}

// 设置并发数
func (this *AsyncEngine) SetLimiter(num int) {
	var limiter = make(chan bool, num)
	this.limiter = limiter
	var chanRequest = make(chan *asyncEngineRequestBody, num)
	this.chanRequest = chanRequest
	var chanRetryRequest = make(chan *asyncEngineRequestBody, num)
	this.chanRetryRequest = chanRetryRequest
	var chanResponses = make(chan *AsyncEngineResponse, num)
	this.ChanResponses = chanResponses
}

func (this *AsyncEngine) Visit(method string, targetUrl string, headers map[string]string, body interface{}, timeout time.Duration, proxies string, meta map[string]interface{}) {
	request := asyncEngineRequestBody{
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
	this.chanRequest <- &request
	go this.get()
}

func (this *AsyncEngine) retryVisit(method string, targetUrl string, headers map[string]string, body interface{}, timeout time.Duration, proxies string, meta map[string]interface{}, retryNumber int64, startTime time.Time) {
	request := asyncEngineRequestBody{
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
	this.chanRetryRequest <- &request
	go this.get()
}

func (this *AsyncEngine) get() {
	this.limiter <- true
	// 获取请求体，优先从重试队列获取
	var request *asyncEngineRequestBody
	if len(this.chanRetryRequest) != 0 {
		request = <-this.chanRetryRequest
	} else {
		request = <-this.chanRequest
	}

	client := http.Client{}
	defer client.CloseIdleConnections()

	// 设置代理和Transport
	addProxyAndTransportErr := this.addProxyAndTransport(&client, request.Proxy, request.Timeout)
	if addProxyAndTransportErr != nil {
		this.onError(nil, addProxyAndTransportErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
		return
	}

	// 如果body不等于nil, 则生成reader类型body
	var payload *strings.Reader
	if request.Body != nil {
		requestBody, isString := request.Body.(string)
		if !isString {
			bodyJson, marshalErr := json.Marshal(request.Body)
			if marshalErr != nil {
				this.onError(nil, marshalErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
				return
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
			this.onError(nil, newRequestErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
			return
		}
	} else {
		req, newRequestErr = http.NewRequest(request.Method, request.URL, payload)
		if newRequestErr != nil {
			this.onError(nil, newRequestErr, request, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
			return
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
		this.onError(res, doErr, request, request.startTime, endTime, consumeTime)
		return
	}
	defer res.Body.Close()
	// 处理返回数据
	this.onResponse(res, request, request.startTime, endTime, consumeTime)
}

func (this *AsyncEngine) onError(res *http.Response, err error, request *asyncEngineRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) {
	var response AsyncEngineResponse
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
	this.ChanResponses <- &response
	<-this.limiter
}

func (this *AsyncEngine) onResponse(res *http.Response, request *asyncEngineRequestBody, startTime time.Time, endTime time.Time, consumeTime float64) {
	var response AsyncEngineResponse
	// 读取响应内容
	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		this.onError(res, err, request, startTime, endTime, consumeTime)
		return
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
	this.ChanResponses <- &response
	<-this.limiter
}

// 实例化异步引擎
func NewAsyncEngine() *AsyncEngine {
	s := AsyncEngine{}
	s.SetLimiter(1)
	s.initTransport()
	return &s
}
