package warner

import (
	"errors"
	"fmt"
	"gogorequest/utils/hash"
	"net/smtp"
	"strings"
	"time"
)

type Email struct {
	User     string
	Password string
	Smtp     string
	Timeout  int64
	Interval map[string]int64
}

func (e *Email) __runSend(to, subject, body, bodyTpye string) error {
	hp := strings.Split(e.Smtp, ":")
	auth := smtp.PlainAuth("", e.User, e.Password, hp[0])

	var contentType string
	if bodyTpye == "html" {
		contentType = "Content-Type: text/" + bodyTpye + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + e.User + "<" + e.User + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	err := smtp.SendMail(e.Smtp, auth, e.User, sendTo, msg)
	return err
}

func (e *Email) Send(to, subject, body, bodyTpye string) error {
	// 将文本内容进行sha1处理
	bodySha1 := hash.StringToSha1(body)
	// 判断此封邮件是否可以现在就发送
	value, ok := e.Interval[bodySha1]
	if ok {
		if time.Now().Unix()-value < e.Timeout {
			return errors.New(fmt.Sprintf("同一封邮件发送频率过高，同样内容的邮件发送间隔时间为: %v秒\n", e.Timeout))
		}
		// 发送邮件
		sendErr := e.__runSend(to, subject, body, bodyTpye)
		if sendErr != nil {
			return sendErr
		}
		e.Interval[bodySha1] = time.Now().Unix()
		return nil
	} else {
		// 发送邮件
		sendErr := e.__runSend(to, subject, body, bodyTpye)
		if sendErr != nil {
			return sendErr
		}
		e.Interval[bodySha1] = time.Now().Unix()
		return nil
	}
}
