package gogorequest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// 异步引擎
type BatchAsyncEngine struct {
	mainEngine // 继承主引擎
}

func (this *BatchAsyncEngine) Visit(targetDatas []BatchAsyncEngineRequestBody) []*BatchAsyncEngineResponse {
	numberOfRequest := len(targetDatas)                      // 请求数
	numberOfResponse := 0                                    // 当前已得到的响应数
	var chanResponses = make(chan *BatchAsyncEngineResponse) // 当前函数作用域的响应队列
	// 分发请求
	for _, targetData := range targetDatas {
		request := batchAsyncEngineRequestBody{
			URL:       targetData.URL,
			Method:    targetData.Method,
			Headers:   targetData.Headers,
			Body:      targetData.Body,
			Proxy:     targetData.Proxy,
			Timeout:   targetData.Timeout,
			Meta:      targetData.Meta,
			startTime: time.Now(),
		}
		go this.get(&request, chanResponses)
	}
	// 处理响应
	result := []*BatchAsyncEngineResponse{} // 此函数return的结果
	for {
		// 如果当前已得到的响应数与请求数相同，说明已获取到全部响应
		if numberOfResponse == numberOfRequest {
			close(chanResponses) // 关闭通道资源
			break
		}
		// 获取响应，将结果插入到返回队列，已得到的响应数+1
		response := <-chanResponses
		result = append(result, response)
		numberOfResponse += 1
	}
	return result
}

func (this *BatchAsyncEngine) get(request *batchAsyncEngineRequestBody, chanResponses chan *BatchAsyncEngineResponse) {
	client := http.Client{}
	defer client.CloseIdleConnections()

	// 设置代理和Transport
	addProxyAndTransportErr := this.addProxyAndTransport(&client, request.Proxy, request.Timeout)
	if addProxyAndTransportErr != nil {
		this.onError(nil, addProxyAndTransportErr, request, chanResponses, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
		return
	}

	// 如果body不等于nil, 则生成reader类型body
	var payload *strings.Reader
	if request.Body != nil {
		requestBody, isString := request.Body.(string)
		if !isString {
			bodyJson, marshalErr := json.Marshal(request.Body)
			if marshalErr != nil {
				this.onError(nil, marshalErr, request, chanResponses, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
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
			this.onError(nil, newRequestErr, request, chanResponses, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
			return
		}
	} else {
		req, newRequestErr = http.NewRequest(request.Method, request.URL, payload)
		if newRequestErr != nil {
			this.onError(nil, newRequestErr, request, chanResponses, request.startTime, time.Now(), time.Now().Sub(request.startTime).Seconds())
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
		this.onError(res, doErr, request, chanResponses, request.startTime, endTime, consumeTime)
		return
	}
	defer res.Body.Close()
	// 处理返回数据
	this.onResponse(res, request, chanResponses, request.startTime, endTime, consumeTime)
}

func (this *BatchAsyncEngine) onError(res *http.Response, err error, request *batchAsyncEngineRequestBody, chanResponses chan *BatchAsyncEngineResponse, startTime time.Time, endTime time.Time, consumeTime float64) {
	var response BatchAsyncEngineResponse
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
	chanResponses <- &response
}

func (this *BatchAsyncEngine) onResponse(res *http.Response, request *batchAsyncEngineRequestBody, chanResponses chan *BatchAsyncEngineResponse, startTime time.Time, endTime time.Time, consumeTime float64) {
	var response BatchAsyncEngineResponse
	// 读取响应内容
	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		this.onError(res, err, request, chanResponses, startTime, endTime, consumeTime)
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
	chanResponses <- &response
}

// 实例化批量异步引擎
func NewBatchAsyncEngine() *BatchAsyncEngine {
	s := BatchAsyncEngine{}
	s.initTransport()
	return &s
}
