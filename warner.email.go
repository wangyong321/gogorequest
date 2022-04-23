package gogogrequest

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"strings"
	"time"
)

type warnerEmail struct {
	user     string
	password string
	smtp     string
	timeout  int64
	interval map[string]int64
}

func (this *warnerEmail) __runSend(to, subject, body, bodyTpye string) error {
	hp := strings.Split(this.smtp, ":")
	auth := smtp.PlainAuth("", this.user, this.password, hp[0])

	var content_type string
	if bodyTpye == "html" {
		content_type = "Content-Type: text/" + bodyTpye + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + this.user + "<" + this.user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(this.smtp, auth, this.user, send_to, msg)
	return err
}

func (this *warnerEmail) __sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func (this *warnerEmail) Send(to, subject, body, bodyTpye string) error {
	// 将文本内容进行sha1处理
	bodySha1 := this.__sha1(body)
	// 判断此封邮件是否可以现在就发送
	value, ok := this.interval[bodySha1]
	if ok {
		if time.Now().Unix()-value < this.timeout {
			return errors.New(fmt.Sprintf("同一封邮件发送频率过高，同样内容的邮件发送间隔时间为: %v秒\n", this.timeout))
		}
		// 发送邮件
		sendErr := this.__runSend(to, subject, body, bodyTpye)
		if sendErr != nil {
			return sendErr
		}
		this.interval[bodySha1] = time.Now().Unix()
		return nil
	} else {
		// 发送邮件
		sendErr := this.__runSend(to, subject, body, bodyTpye)
		if sendErr != nil {
			return sendErr
		}
		this.interval[bodySha1] = time.Now().Unix()
		return nil
	}
}
