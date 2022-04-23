## 使用方法

### 安装

```shell
go get -u github.com/wangyong321/gogorequest@v1.0.1
```

### 同步下载引擎

```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogogrequest.NewSyncEngine()
	resp := s.Visit("GET", "https://httpbin.org/get", nil, nil, 10, "", nil)
	fmt.Println(resp.Text)
}
```

### 流式并发请求

```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogogrequest.NewAsyncEngine()
	s.SetLimiter(10)
	go func() {
		for {
			s.Visit("GET", "https://httpbin.org/get", nil, nil, 5, "", nil)
		}
	}()

	for {
		resp := <-s.ChanResponses
		fmt.Println(resp.Text)
	}
}
```

### 批量并发请求

```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogogrequest.NewBatchAsyncEngine()
	targetDatas := []gogogrequest.BatchAsyncEngineRequestBody{}
	// 批量生成任务
	for i := 1; i <= 5; i++ {
		var request gogogrequest.BatchAsyncEngineRequestBody
		request.URL = "https://httpbin.org/get"
		request.Method = "GET"
		request.Headers = nil
		request.Body = nil
		request.Timeout = 10
		request.Proxy = ""
		request.Meta = nil
		targetDatas = append(targetDatas, request)
	}
	// 请求
	resps := s.Visit(targetDatas)
	for index, resp := range resps {
		fmt.Printf("%d. %v\n", index+1, resp.Text)
	}
}
```

### 请求重试

```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	// 最大重试次数
	maxRetryCount := 3
	s := gogogrequest.NewAsyncEngine()
	go func() {
		for {
			s.Visit("GET", "https://httpbin.org/get", nil, nil, 10, "", nil)
		}
	}()
	for {
		resp := <-s.ChanResponses
		if resp.Error != nil {
			if resp.Request.RetryNumber == int64(maxRetryCount) {
				// 如果当前重试次数等于最大重试次数要求，则放弃重试
				continue
			} else {
				// 开始重试重试请求
				resp.Request.Retry()
				continue
			}
		}
		fmt.Println(resp.Text)
	}
}

```

### 发送飞书消息

```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogogrequest.NewBatchAsyncEngine()                    // 任意引擎
	api := "https://open.feishu.cn/open-apis/bot/v2/hook/demo" // 飞书机器人api
	token := "demo"                                            // 飞书机器人api请求token
	msg := "本条消息由go语言测试程序发出"                                   // 需要飞书机器人发送的消息
	timeout := 10                                              // 同一条消息的发送时间间隔限制，防止暴力发送

	s.OpenFeiShuWarner(api, token, int64(timeout)) // 引擎开启飞书预警消息模块
	body, err := s.WarnerFeiShu.Send(msg)          // 发送飞书消息
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
}
```

### 发送邮件消息

```go
package main

import (
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogogrequest.NewBatchAsyncEngine()                                  // 任意引擎
	fromEmail := "demo@163.com"                                              // 发送邮件的邮箱
	fromPassword := "HDIDWUWTAJUIJPMW"                                       // 发送邮件的邮箱密码
	fromSmtp := "smtp.163.com:25"                                            // 发送邮件的smtp
	timeout := 10                                                            // 同一内容邮件的发送时间间隔限制，防止暴力发送
	s.OpenEmailWarner(fromEmail, fromPassword, fromSmtp, int64(timeout))     // 引擎开启邮件预警消息模块
	err := s.WarnerEmail.Send("demo@163.com", "ceshi youjian", "自动测试邮件", "") // 发送普通文本
	//err := s.WarnerEmail.Send("wy@3ydata.com", "ceshi youjian", "<h1>自动测试邮件</h1>", "html") // 发送html文本
	if err != nil {
		panic(err)
	}
}
```