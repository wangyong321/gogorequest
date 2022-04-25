package gogorequest

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/url"
	"time"
)

// 主引擎
type mainEngine struct {
	transport    *http.Transport
	WarnerEmail  *warnerEmail
	WarnerFeiShu *warnerFeiShu
}

// 初始化默认transport
func (this *mainEngine) initTransport() {
	// 默认transport
	transport := http.Transport{
		DialContext:           (&net.Dialer{}).DialContext,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,              // 最大空闲连接数
		IdleConnTimeout:       60 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout:   60 * time.Second, // TLS 握手超时
		ExpectContinueTimeout: 1 * time.Second,
	}
	this.transport = &transport
}

// 读取证书
func ReadCrt(cerPath string, keyPath string) (*tls.Config, error) {
	cliCrt, err := tls.LoadX509KeyPair(cerPath, keyPath)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cliCrt},
	}
	return tlsConfig, nil
}

// 重置自定义transport
func (this *mainEngine) SetTransport(transport *http.Transport) {
	this.transport = transport
}

// 开启HTTP2请求
func (this *mainEngine) EnableHTTP2() error {
	err := http2.ConfigureTransport(this.transport)
	if err != nil {
		return err
	}
	return nil
}

// 为请求client设置代理IP、transport、请求超时
func (this *mainEngine) addProxyAndTransport(client *http.Client, proxies string, timeout time.Duration) error {
	if proxies != "" {
		proxyUrl, err := url.Parse(proxies)
		if err != nil {
			return err
		}
		this.transport.Proxy = http.ProxyURL(proxyUrl)
	}
	client.Transport = this.transport
	client.Timeout = timeout * time.Second
	return nil
}

// 开启邮件报警器
func (this *mainEngine) OpenEmailWarner(user, password, smtp string, timeout int64) {
	if timeout < 10 {
		panic("邮件预警的消息发送间隔不能小于10秒")
	}
	this.WarnerEmail = &warnerEmail{
		user:     user,
		password: password,
		smtp:     smtp,
		timeout:  timeout,
	}
	this.WarnerEmail.interval = map[string]int64{}
}

// 开启飞书报警器
func (this *mainEngine) OpenFeiShuWarner(api, token string, timeout int64) {
	if timeout < 10 {
		panic("飞书预警的消息发送间隔不能小于10秒")
	}
	this.WarnerFeiShu = &warnerFeiShu{
		api:     api,
		token:   token,
		timeout: timeout,
	}
	this.WarnerFeiShu.interval = map[string]int64{}
}
