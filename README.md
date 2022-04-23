## 使用方法

```go
func main() {
    startTime := time.Now().Unix()
    // 同步爬虫
    s := gogrequest.NewSyncEngine()
    resp := s.Visit("GET", targetUrl, nil, nil, 5, "", nil)
    endTime := time.Now().Unix()
    fmt.Println(resp.Text)
    fmt.Printf("耗时: %v秒\n", endTime-startTime)

    //// 异步爬虫
    //s := gogrequest.NewAsyncEngine()
    //s.SetLimiter(10)
    //go func() {
    //	for {
    //	    s.Visit("GET", targetUrl, nil, nil, 5, "", nil)
    //	}
    //}()
    //for {
    //	resp := <-s.ChanResponses
    //	endTime := time.Now().Unix()
    //	fmt.Println(resp.Text)
    //	fmt.Printf("耗时: %v秒\n", endTime-startTime)
    //}
	
	//// 批量异步爬虫
    //s := gogrequest.NewBatchAsyncEngine()
    //for {
    //    // 1. 生成批量请求数组
    //    targetDatas := []gogrequest.BatchAsyncEngineRequestBody{}
    //    for i := 1; i <= 5; i++ {
    //        var request gogrequest.BatchAsyncEngineRequestBody
    //        request.URL = "http://127.0.0.1:9091/"
    //        request.Method = "GET"
    //        request.Headers = nil
    //        request.Body = nil
    //        request.Timeout = 10
    //        request.Proxy = ""
    //        request.Meta = nil
    //        targetDatas = append(targetDatas, request)
    //    }
    //    
    //    fmt.Printf("当前请求数: %d\n", len(targetDatas))
    //    // 2. 开始批量请求
    //    resps := s.Visit(targetDatas)
    //    // 3. 处理批量响应
    //    for index, resp := range resps {
    //        fmt.Printf("%d. %v\n", index+1, resp)
    //    }
    //    fmt.Printf("当前响应数: %d\n", len(resps))
    //    time.Sleep(3 * time.Second)
    //}
	
	//// 使用爬虫引擎发送飞书消息
    //api := "https://open.feishu.cn/open-apis/bot/v2/hook/demo"                                 // 飞书机器人api
    //token := "demo"                                                                            // 飞书机器人api请求token
    //msg := "本条消息由go语言测试程序发出"                                                          // 需要飞书机器人发送的消息
    //timeout := 10                                                                              // 同一条消息的发送时间间隔限制，防止暴力发送
    //s := gogrequest.NewBatchAsyncEngine()                                                      // 实例化爬虫引擎[任意引擎均可]
    //s.OpenFeiShuWarner(api, token, int64(timeout))                                             // 引擎开启飞书预警消息模块
    //body, err := s.WarnerFeiShu.Send(msg)                                                      // 发送飞书消息
    //if err != nil {
    //    panic(err)
    //}
    //fmt.Println(body)
	
	//// 使用爬虫引擎发送邮件消息
    //fromEmail := "demo@163.com"                                               // 发送邮件的邮箱
    //fromPassword := "HDIDWUWTAJUIJPMW"                                        // 发送邮件的邮箱密码
    //fromSmtp := "smtp.163.com:25"                                             // 发送邮件的smtp
    //timeout := 10                                                             // 同一内容邮件的发送时间间隔限制，防止暴力发送
    //s := gogrequest.NewBatchAsyncEngine()                                     // 实例化爬虫引擎[任意引擎均可]
    //s.OpenEmailWarner(fromEmail, fromPassword, fromSmtp, int64(timeout))      // 引擎开启邮件预警消息模块
    //err := s.WarnerEmail.Send("demo@163.com", "ceshi youjian", "自动测试邮件", "") // 发送普通文本
    ////err := s.WarnerEmail.Send("wy@3ydata.com", "ceshi youjian", "<h1>自动测试邮件</h1>", "html") // 发送html文本
    //if err != nil {
    //    panic(err)
    //}
}
```