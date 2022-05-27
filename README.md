## 使用方法

### 安装

```shell
go get -u github.com/wangyong321/gogorequest
```

### 同步下载引擎

```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogorequest.NewSyncEngine()
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
	s := gogorequest.NewAsyncEngine()
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
	s := gogorequest.NewBatchAsyncEngine()
	targetDatas := []gogorequest.BatchAsyncEngineRequestBody{}
	// 批量生成任务
	for i := 1; i <= 5; i++ {
		var request gogorequest.BatchAsyncEngineRequestBody
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

### 流式下载文件
```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogorequest.NewFileEngine()
	targetUrl := "https://ip3875670905.mobgslb.tbcache.com/fs08/2022/05/06/11/109_593a72f6ef92b3220c052a744d11dc08.apk?yingid=wdj_web&fname=%E6%A2%A6%E5%B9%BB%E8%A5%BF%E6%B8%B8&productid=2011&pos=wdj_web%2Fdetail_normal_dl%2F0&appid=6602792&packageid=100724749&apprd=6602792&iconUrl=http%3A%2F%2Fandroid-artworks.25pp.com%2Ffs08%2F2022%2F05%2F07%2F5%2F109_61dd3fd76244facbb759fb2682b0c196_con.png&pkg=com.netease.my.uc&did=d16f06fab6bce8562f398fb0899b4790&vcode=13600&md5=aa5fc49fd2f40addb899b51e772841f2&ali_redirect_domain=alissl.ucdl.pp.uc.cn&ali_redirect_ex_ftag=a2101f161c398cbf3a62935698fddb4147583998c6c0062b&ali_redirect_ex_tmining_ts=1653660871&ali_redirect_ex_tmining_expire=3600&ali_redirect_ex_hot=100"
	headers := map[string]string{
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language":           "zh-CN,zh;q=0.9",
	}
	resp := s.Visit("GET", targetUrl, headers, nil, -1, "", fmt.Sprintf("梦幻西游.pkg"))
	if resp.Error != nil {
		panic(resp.Error)
	}
	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return
	}
	fmt.Println(resp.Text)
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
	s := gogorequest.NewAsyncEngine()
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

### 开启HTTP2.0模式
```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
)

func main() {
	s := gogorequest.NewSyncEngine()
	s.EnableHTTP2()
	resp := s.Visit("GET", "https://httpbin.org/get", nil, nil, 10, "", nil)
	fmt.Println(resp.Text)
}
```

### 挂载证书请求
```go
package main

import (
	"fmt"
	"github.com/wangyong321/gogorequest"
	"net"
	"net/http"
	"time"
)

func main() {
	s := gogorequest.NewSyncEngine()
	pemPath := "ca.pem"
	keyPath := "ca.key"
	// 请求器配置TLS证书
	tlsConfig, err := s.ReadCrt(pemPath, keyPath)
	if err != nil {
		panic(err)
	}
	// Transport指定TLS证书
	transport := http.Transport{
		DialContext:           (&net.Dialer{}).DialContext,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,              // 最大空闲连接数
		IdleConnTimeout:       60 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout:   60 * time.Second, // TLS 握手超时
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig, // 加载TLS
	}
	// 重置请求器transport
	s.SetTransport(&transport)
	// 请求
	resp := s.Visit("GET", "https://httpbin.org/get", nil, nil, 10, "", nil)
	fmt.Println(resp.Text)
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
	s := gogorequest.NewBatchAsyncEngine()                    // 任意引擎
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
	s := gogorequest.NewBatchAsyncEngine()                                  // 任意引擎
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