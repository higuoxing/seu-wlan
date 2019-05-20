package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const SEU_WLAN_LOGIN_URL = "https://w.seu.edu.cn/index.php/index/login"

// Loggers
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Command line options
type Options struct {
	config                 string
	username               string
	password               string
	interval               int
	enableMacAuth          bool
	disableTLSVerification bool
}

var options *Options

func loggerInit() {
	Info = log.New(os.Stdout, "[Info]    ", log.Ldate|log.Ltime)
	Warning = log.New(os.Stdout, "[Warning] ", log.Ldate|log.Ltime)
	Error = log.New(os.Stdout, "[Error]   ", log.Ldate|log.Ltime)
}

func init() {
	// init command options parser
	options = &Options{}

	flag.StringVar(&options.username, "u", "", "Your card number. (Required)")
	flag.StringVar(&options.password, "p", "", "Your password. (Required)")
	flag.StringVar(&options.config, "c", "", "Your config file.")
	flag.IntVar(&options.interval, "i", 0, "Run this tool periodically.")
	flag.BoolVar(&options.enableMacAuth, "enable-mac-auth", false, "Enable this machine's mac address to be remembered.")
	flag.BoolVar(&options.disableTLSVerification, "disable-tls-verification", false, "Disable TLS certificate verification.")

	flag.Usage = func() {
		fmt.Println("Usage: seu-wlan [options] param")
		flag.PrintDefaults()
	}

	// init loggers
	loggerInit()
}

func main() {
	flag.Parse()

	err := checkOptions(options)
	if err != nil {
		Error.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	param := encodeParam(options)

	if options.interval > 0 {
		runInLoop(param, options.interval)
	} else {
		runOnce(param)
	}
	return
}

func encodeParam(options *Options) url.Values {

	b64pass := base64.StdEncoding.EncodeToString([]byte(options.password))

	var macAuth string

	if options.enableMacAuth {
		macAuth = "1"
	} else {
		macAuth = "0"
	}

	return url.Values{
		"username":      {options.username},
		"password":      {string(b64pass)},
		"enablemacauth": {macAuth},
	}
}

func loginRequest(param url.Values, interval int) (error, map[string]interface{}) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: options.disableTLSVerification},
	}

	var client *http.Client

	if interval != 0 {
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * time.Duration(interval),
		}
	} else {
		client = &http.Client{
			Transport: tr,
		}
	}
	response, err := client.PostForm(SEU_WLAN_LOGIN_URL, param)
	if err != nil {
		return fmt.Errorf("HTTP Request Error: %s", "error occurred when sending post request"), nil
	}
	defer response.Body.Close()

	loginMsgRaw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Read Response Error: %s", "error occurred when reading response from server"), nil
	}

	var loginMsgJson map[string]interface{}
	err = json.Unmarshal(loginMsgRaw, &loginMsgJson)
	if err != nil {
		return fmt.Errorf("Parsing JSON Error: %s", "error occurred when parsing JSON format response"), nil
	}

	return nil, loginMsgJson
}

func emitLog(loginMsgJson map[string]interface{}) {
	if loginMsgJson["status"] == 1.0 {
		Info.Printf("%v\tlogin user: %v\tlogin ip: %v\tlogin loc: %v\n",
			loginMsgJson["info"],
			loginMsgJson["logout_username"],
			loginMsgJson["logout_ip"],
			loginMsgJson["logout_location"])
	} else {
		Info.Println(loginMsgJson["info"])
	}
}

func runInLoop(param url.Values, interval int) {
	for {
		err, loginMsgJson := loginRequest(param, interval)
		if err != nil {
			Error.Println(err)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		emitLog(loginMsgJson)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func runOnce(param url.Values) {
	err, loginMsgJson := loginRequest(param, 0)
	if err != nil {
		Error.Println(err)
		return
	}
	emitLog(loginMsgJson)
}

func checkOptions(options *Options) error {
	if options.config != "" {
		/* read from config file */
		err := readConfigFile(options.config, options)
		if err != nil {
			return err
		}
	}

	if options.username == "" || options.password == "" {
		return fmt.Errorf("Command Parsing Error: %s", "username and password are required")
	} else if options.interval < 0 {
		return fmt.Errorf("Command Parsing Error: %s", "-i option cannot be less than 0")
	}

	return nil
}

func readConfigFile(path string, options *Options) error {
	jsonFile, err := os.Open(path)
	defer jsonFile.Close()

	if err != nil {
		return fmt.Errorf("Config File Parsing Error: %s", "cannot find your config file in this path")
	}

	byteVal, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		return fmt.Errorf("Config File Parsing Error: %s", "an error occurred when reading config file")
	}

	var configJson map[string]interface{}
	err = json.Unmarshal(byteVal, &configJson)

	if err != nil {
		return fmt.Errorf("Config File Parsing Error: %s", "an error occurred when parsing config file")
	}

	if configJson["username"] == nil || configJson["password"] == nil {
		return fmt.Errorf("Config File Parsing Error: %s", "username and password are required")
	}

	switch ty := configJson["username"].(type) {
	default:
		return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("username should be string format, not %T", ty))
	case string:
		options.username = configJson["username"].(string)
	}

	switch ty := configJson["password"].(type) {
	default:
		return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("password should be string format, not %T", ty))
	case string:
		options.password = configJson["password"].(string)
	}

	// optional
	if configJson["interval"] != nil {
		switch ty := configJson["interval"].(type) {
		default:
			return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("interval should be integer, not %T", ty))
		case float64:
			options.interval = int(configJson["interval"].(float64))
		}
	}

	// optional
	if configJson["enable-mac-auth"] != nil {
		switch ty := configJson["enable-mac-auth"].(type) {
		default:
			return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("enable-mac-auth should be boolean, not %T", ty))
		case bool:
			options.enableMacAuth = bool(configJson["enable-mac-auth"].(bool))
		}
	}

	// optional
	if configJson["disable-tls-verification"] != nil {
		switch ty := configJson["disable-tls-verification"].(type) {
		default:
			return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("disable-tls-verification should be boolean, not %T", ty))
		case bool:
			options.disableTLSVerification = bool(configJson["disable-tls-verification"].(bool))
		}
	}

	return nil
}
