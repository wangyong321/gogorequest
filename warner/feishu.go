package warner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"gogorequest/utils/hash"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type FeiShu struct {
	Api      string
	Token    string
	Timeout  int64
	Interval map[string]int64
}

func (fs *FeiShu) __createSign(timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + fs.Token
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func (fs *FeiShu) __runSend(bodySha1 string, msg string) (string, error) {
	timestamp := time.Now().Unix()
	sign, signErr := fs.__createSign(timestamp)
	if signErr != nil {
		return "", signErr
	}
	requestBody := fmt.Sprintf(`{"timestamp": %d,"sign": "%vs","msg_type": "text","content": {"text": "%vs"}}`, timestamp, sign, msg)
	payload := strings.NewReader(requestBody)
	client := &http.Client{}
	req, newReqErr := http.NewRequest("POST", fs.Api, payload)

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
	fs.Interval[bodySha1] = time.Now().Unix()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (fs *FeiShu) Send(msg string) (string, error) {
	// 将文本内容进行sha1处理
	bodySha1 := hash.StringToSha1(msg)
	// 判断此封邮件是否可以现在就发送
	value, ok := fs.Interval[bodySha1]
	if ok {
		if time.Now().Unix()-value < fs.Timeout {
			return "", errors.New(fmt.Sprintf("同一条信息发送频率过高，同样内容的消息发送间隔时间为: %v秒\n", fs.Timeout))
		}
		return fs.__runSend(bodySha1, msg)
	} else {
		fs.Interval[bodySha1] = time.Now().Unix()
		return fs.__runSend(bodySha1, msg)
	}
}
