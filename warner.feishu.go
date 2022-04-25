package gogorequest

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type warnerFeiShu struct {
	api      string
	token    string
	timeout  int64
	interval map[string]int64
}

func (this *warnerFeiShu) __createSign(timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + this.token
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func (this *warnerFeiShu) __sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func (this *warnerFeiShu) __runSend(bodySha1 string, msg string) (string, error) {
	timestamp := time.Now().Unix()
	sign, signErr := this.__createSign(timestamp)
	if signErr != nil {
		return "", signErr
	}
	requestBody := fmt.Sprintf(`{"timestamp": %d,"sign": "%s","msg_type": "text","content": {"text": "%s"}}`, timestamp, sign, msg)
	payload := strings.NewReader(requestBody)
	client := &http.Client{}
	req, newReqErr := http.NewRequest("POST", this.api, payload)

	if newReqErr != nil {
		return "", newReqErr
	}
	req.Header.Add("Content-Type", "application/json")

	res, doErr := client.Do(req)
	if doErr != nil {
		return "", doErr
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	this.interval[bodySha1] = time.Now().Unix()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (this *warnerFeiShu) Send(msg string) (string, error) {
	// 将文本内容进行sha1处理
	bodySha1 := this.__sha1(msg)
	// 判断此封邮件是否可以现在就发送
	value, ok := this.interval[bodySha1]
	if ok {
		if time.Now().Unix()-value < this.timeout {
			return "", errors.New(fmt.Sprintf("同一条信息发送频率过高，同样内容的消息发送间隔时间为: %v秒\n", this.timeout))
		}
		return this.__runSend(bodySha1, msg)
	} else {
		this.interval[bodySha1] = time.Now().Unix()
		return this.__runSend(bodySha1, msg)
	}
}
