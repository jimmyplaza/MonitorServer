package main
import(
	"fmt"
	"log/syslog"
	"crypto/rc4"
)

type SyslogSender struct {
	Writer *syslog.Writer
	key []byte // min 1 byte, max is 256 bytes
}

func (sl *SyslogSender) Encrypt(message string) string {
	c, err := rc4.NewCipher(sl.key)
	if err != nil {
	    // this is the KeySizeError output if the key is less than 1 byte
	    // or more than 256 bytes
	    fmt.Println(err.Error) // KeySizeError
	    return err.Error()
    }

  	plaintext := []byte(message)

  	//encrypt
  	encrypted := make([]byte, len(plaintext))
  	c.XORKeyStream(encrypted, plaintext)
  	fmt.Printf("[%s] encrypted to [%x] by rc4 crypto\n", plaintext, encrypted)
  	c.Reset() // reset the key data for clean data
	return fmt.Sprintf("%x",encrypted)
}

func (sl *SyslogSender) Decrypt (encrypted []byte) string {
	//decrypt
	decrypted := make([]byte, len(encrypted))
	
	c, err := rc4.NewCipher(sl.key)
	if err != nil {
	   fmt.Println(err.Error)
	   return ""
	}
	
	c.XORKeyStream(decrypted, encrypted)
	c.Reset() // reset the key data for clean data
	fmt.Printf("[%x] decrypted to [%s] \n", encrypted, decrypted)
	return string(decrypted)
}

func (sl *SyslogSender) Write(network, raddr string,priority int,tag string,logMessage string) {
	var	pri syslog.Priority
	switch priority {
		case 8:
			pri = syslog.LOG_EMERG
		case 7:
			pri = syslog.LOG_ALERT
		case 6:
			pri = syslog.LOG_CRIT
		case 5:
			pri = syslog.LOG_ERR
		case 4:
			pri = syslog.LOG_WARNING
		case 3:
			pri = syslog.LOG_NOTICE
		case 2:
			pri = syslog.LOG_INFO
		default:
			pri = syslog.LOG_DEBUG		
	}
	println("priority:",priority,pri)
	l2c, err := syslog.Dial(network, raddr, pri, tag) // connection to a log daemon
	defer l2c.Close()
	if err != nil {
		fmt.Println("error",err)
	}
	l2c.Write([]byte(sl.Encrypt(logMessage)))
}

/*
func main(){
	sl := &SyslogSender{key:[]byte("cert")}
	sl.Write("udp", "localhost:7900", 1,"SyslogSender","test syslog")
}
*/
