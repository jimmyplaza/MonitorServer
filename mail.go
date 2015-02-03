package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"
	"gopkg.in/gomail.v1"
)

/*
var (
	smtpAccount = struct {
		Email    string
		Password string
	}{
		"g2.service@nexusguard.com",
		"NxGdEv139",
	}
)
*/
//SendHTMLMail send HTML document to multiple emails, with exist smtp configuration
func SendHTMLMail(SmtpServer, Port, From string, Too []string, Title, BodyMsg string) {
	port, err := strconv.Atoi(cfg.Mail.Port)

	if err != nil {
		log.Println("Smtp mail serer not able convert integer", Port)
		return
	}

	//mailer := gomail.NewMailer(SmtpServer, smtpAccount.Email, smtpAccount.Password, port)
	mailer := gomail.NewMailer(cfg.Mail.SmtpServer, cfg.Mail.From, cfg.Mail.Password, port)

	// FIXME this should move out off this block
	BodyMsg = Title + "<br>" + BodyMsg

	v := url.Values{}
	v.Set("from", "g2.service@nexusguard.com")
	v.Set("subject", Title)
	v.Set("content", BodyMsg)

	msg := gomail.NewMessage()
	msg.SetHeader("From", cfg.Mail.From)
	msg.SetHeader("To", strings.Join(Too, ","))
	msg.SetHeader("Subject", Title)
	msg.SetBody("text/html", BodyMsg)

	if err := mailer.Send(msg); err != nil {
		log.Println(err)
	}

}

func MorningMail(SmtpServer, Port, From string, Too []string, Title, BodyMsg string) {
	var To string
	for _, t := range Too {
		To = To + t + " , "
	}
	To = To[:len(To)-3]
	//uurl := "http://g2tool.cloudapp.net:445/morningbird"
	uurl := "http://gcptools.nexusguard.com:2999/morningbird"
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
		},
	}
	Title = Title
	BodyMsg = Title + "<br>" + BodyMsg

	v := url.Values{}
	v.Set("to", To)
	v.Set("from", "g2.service@nexusguard.com")
	v.Set("subject", Title)
	v.Set("content", BodyMsg)
	v.Set("publickey", "cba2eb")
	v.Set("privatekey", "c3e12e")

	//out, _ := json.Marshal(m)
	//outReader := bytes.NewReader([]byte(out))
	//res, err := myClient.Post(uurl, "application/x-www-form-urlencoded", outReader)
	//res, err := myClient.Post(url, "application/json", outReader)
	//res, err := myClient.PostForm(uurl, url.Values{ "from" : { "g2.service@nexusguard.com" }, "to" : { "jimmy.ko@nexusguard.com, stickbob@gmail.com"  }, "subject" : {"aaaa"}, "content":{"ttttt"}, "publickey":{"cba2eb"}, "privatekey":{"c3e12e"}  })
	res, err := myClient.PostForm(uurl, v)
	if err != nil {
		fmt.Printf("MorningBird Mail Error:%s\n", err)
		return
	}
	if res.StatusCode != 200 {
		fmt.Printf("MorningBird Mail Error code: %d,url:%s\n", res.StatusCode, uurl)
	}
	//err = json.Unmarshal([]byte(res.Body), &obj)
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Read Body Error:%s\n", err)
		//errMsg := fmt.Sprintf("%s",err)
		//WriteToSyslog(3,"Monitor-MorningMail",errMsg)
		res.Body.Close()
	}
	var obj interface{}
	err = json.Unmarshal(contents, &obj)
	if *debug {
		fmt.Println(obj)
	}
	if err != nil {
		//errMsg := fmt.Sprintf("%s",err)
		fmt.Printf("MorningMail JSON Error:%s => %s\n", uurl, err)
		//WriteToSyslog(3,"Monitor-MorningMail",errMsg)
	}
	res.Body.Close()
}

func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <>")
}

func SendMailSSL(host string, port uint, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {
	serverAddr := fmt.Sprintf("%s:%d", host, port)
	//serverAddr := fmt.Sprintf("%s",host)

	conn, err := tls.Dial("tcp", serverAddr, nil)
	if err != nil {
		log.Println("Error Dialing", err)
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Println("Error SMTP connection", err)
		return err
	}

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if len(addr) > 1 {
			if err = c.Rcpt(addr); err != nil {
				return err
			}
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

func SendMail(SmtpServer string, Port uint, From string, To string, Url string, errMsg string, rspStatus string) {
	// Set up authentication information.
	smtpServer := SmtpServer

	//smtpServer := "smtp.gmail.com"
	auth := smtp.PlainAuth(
		"",
		From, //"stickbob@gmail.com"
		"y2kjimmy",
		smtpServer,
	)

	//from := mail.Address{"Monitor Center", "stickbob@gmail.com"}
	//to := mail.Address{"Recipient", "jimmy.ko@nexusguard.com"}
	from := mail.Address{"", From}
	//to := mail.Address{"Recipient", To}
	to := mail.Address{"", To}
	//to := mail.Address{"Recipient", "stickbob@gmail.com"}
	//title := "[" + Url + "]" + " status: stop "
	title := "[g2Monitor]" + " - [" + Url + "]" + " status: stop "
	body := "[" + Url + "]" + " is down\n" + "STATUS CODE: " + rspStatus + "\nERROR: " + errMsg

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = encodeRFC2047(title)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	/*	err := smtp.SendMail(
			//smtpServer + ":465",
			smtpServer,
			auth,
			from.Address,
			[]string{to.Address},
			[]byte(message),
			//[]byte("This is the email body."),
		)
	*/
	//host string, port uint, auth smtp.Auth, from string,
	//to []string, msg []byte)
	var To_list = make([]string, 0)
	arr := strings.Split(To, ",")
	for _, v := range arr {
		v = strings.TrimSpace(v)
		if len(v) > 1 {
			To_list = append(To_list, v)
		}
	}

	err := SendMailSSL(
		smtpServer,
		Port,
		auth,
		from.Address,
		//[]string{to.Address,"stickclinton100@hotmail.com"},
		To_list,
		[]byte(message),
	)
	if err != nil {
		//log.Fatal(err)
		fmt.Println("MAILFATAL ERROR")
		//WriteToLogFile(Url, "MAILFATAL ERROR", "", filepath1)
	}
}
