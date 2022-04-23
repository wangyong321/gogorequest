package engine

import (
	"gogorequest/warner"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/url"
	"time"
)

// 主引擎被各专项引擎所继承
type mainEngine struct {
	transport    *http.Transport
	WarnerEmail  *warner.Email
	WarnerFeiShu *warner.FeiShu
}

// 初始化默认transport
func (m *mainEngine) initTransport() {
	// 默认transport
	transport := http.Transport{
		DialContext:           (&net.Dialer{}).DialContext,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,              // 最大空闲连接数
		IdleConnTimeout:       60 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout:   60 * time.Second, // TLS 握手超时
		ExpectContinueTimeout: 1 * time.Second,
	}
	m.transport = &transport
}

// SetTransport 重置自定义transport
func (m *mainEngine) SetTransport(transport *http.Transport) {
	m.transport = transport
}

// EnableHTTP2 开启HTTP2请求
func (m *mainEngine) EnableHTTP2() error {
	err := http2.ConfigureTransport(m.transport)
	if err != nil {
		return err
	}
	return nil
}

// 为请求client设置代理IP、transport、请求超时
func (m *mainEngine) addProxyAndTransport(client *http.Client, proxies string, timeout time.Duration) error {
	if proxies != "" {
		proxyUrl, err := url.Parse(proxies)
		if err != nil {
			return err
		}
		m.transport.Proxy = http.ProxyURL(proxyUrl)
	}
	client.Transport = m.transport
	client.Timeout = timeout * time.Second
	return nil
}

// OpenEmailWarner 开启邮件报警器
func (m *mainEngine) OpenEmailWarner(user, password, smtp string, timeout int64) {
	if timeout < 10 {
		panic("邮件预警的消息发送间隔不能小于10秒")
	}
	m.WarnerEmail = &warner.Email{
		User:     user,
		Password: password,
		Smtp:     smtp,
		Timeout:  timeout,
	}
	m.WarnerEmail.Interval = map[string]int64{}
}

// OpenFeiShuWarner 开启飞书报警器
func (m *mainEngine) OpenFeiShuWarner(api, token string, timeout int64) {
	if timeout < 10 {
		panic("飞书预警的消息发送间隔不能小于10秒")
	}
	m.WarnerFeiShu = &warner.FeiShu{
		Api:     api,
		Token:   token,
		Timeout: timeout,
	}
	m.WarnerFeiShu.Interval = map[string]int64{}
}
