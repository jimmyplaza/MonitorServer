package main 

import (
	"fmt"
	"net/mail" 
	"strings"
	"net/smtp"
	"log"
	"encoding/base64"
	"crypto/tls"
	
)



func encodeRFC2047(String string) string{
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
    	if len(addr) > 1{
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


func SendMail(SmtpServer string, Port uint, From string, To string, Url string, errMsg string, rspStatus string){
	// Set up authentication information.
 	smtpServer := SmtpServer

	//smtpServer := "smtp.gmail.com"
	auth := smtp.PlainAuth(
		"",
		From,  //"stickbob@gmail.com"
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
	title := "[g2Monitor]" +  " - [" + Url + "]" + " status: stop "
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
	arr := strings.Split(To,",")
	for _, v := range arr {
		v  = strings.TrimSpace(v)
		if len(v) > 1{
			To_list = append(To_list,v)
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
