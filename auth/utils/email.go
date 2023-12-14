package utils

import (
	"crypto/tls"
	"math/rand"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

func SendMail(toEmail, title, content string) error {
	// QQ 邮箱：
	// SMTP 服务器地址：smtp.qq.com（SSL协议端口：465/994 | 非SSL协议端口：25）
	host := "smtp.qq.com"
	port := 25
	userName := "635969862@qq.com"
	password := "xkyyzqitpbrjbbbg"

	m := gomail.NewMessage()
	m.SetHeader("From", userName) // 发件人
	m.SetHeader("To", toEmail)    // 收件人，可以多个收件人，但必须使用相同的 SMTP 连接
	m.SetHeader("Subject", title) // 邮件主题

	// text/html 的意思是将文件的 content-type 设置为 text/html 的形式，浏览器在获取到这种文件时会自动调用html的解析器对文件进行相应的处理。
	// 可以通过 text/html 处理文本格式进行特殊处理，如换行、缩进、加粗等等
	m.SetBody("text/html", content)

	// text/plain的意思是将文件设置为纯文本的形式，浏览器在获取到这种文件时并不会对其进行处理
	// m.SetBody("text/plain", "纯文本")
	// m.Attach("test.sh")   // 附件文件，可以是文件，照片，视频等等
	// m.Attach("lolcatVideo.mp4") // 视频
	// m.Attach("lolcat.jpg") // 照片
	d := gomail.NewDialer(
		host,
		port,
		userName,
		password,
	)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}

func RandomCount(count int) string {
	str := ""
	rand.Seed(time.Now().UnixMicro())
	for i := 0; i < count; i++ {
		str = str + strconv.Itoa(rand.Intn(10))
	}

	return str
}
