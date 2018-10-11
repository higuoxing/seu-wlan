package main

import (
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

var SEU_WLAN_LOGIN_URL = "http://w.seu.edu.cn/index.php/index/login"

// Loggers
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Command line options
type Options struct {
	username string
	password string
	macauth  int
	interval int
}

var options *Options

// Runtime error
type RuntimeError struct {
	errType string
	errHint string
}

func (err *RuntimeError) Error() string {
	return fmt.Sprintf("[%v]  %v", err.errType, err.errHint)
}

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
	flag.IntVar(&options.macauth, "m", 0, "Enable seu-wlan remember your mac address. 0 (default) or 1.")
	flag.IntVar(&options.interval, "i", 0, "Enable this plugin run in loop and request seu-wlan login server.")
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
		err := runInLoop(param, options.interval)
		if err != nil {
			Error.Println(err)
			os.Exit(1)
		}
	} else {
		err := runOnce(param)
		if err != nil {
			Error.Println(err)
			os.Exit(1)
		}
	}
}

func encodeParam(options *Options) url.Values {
	b64pass := base64.StdEncoding.EncodeToString([]byte(options.password))
	return url.Values{"username": {options.username},
		"password":      {string(b64pass)},
		"enablemacauth": {string(options.macauth)}}
}

func loginRequest(param url.Values, interval int) (error, map[string]interface{}) {
	var client *http.Client
	if interval > 0 {
		client = &http.Client{Timeout: time.Second * time.Duration(interval)}
	} else {
		client = &http.Client{}
	}
	response, err := client.PostForm(SEU_WLAN_LOGIN_URL, param)
	if err != nil {
		return &RuntimeError{"HTTP Request Error", "error occurred when sending post request"}, nil
	}
	defer response.Body.Close()

	loginMsgRaw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &RuntimeError{"Read Response Error", "error occurred when reading response from server"}, nil
	}

	var loginMsgJson map[string]interface{}
	err = json.Unmarshal(loginMsgRaw, &loginMsgJson)
	if err != nil {
		return &RuntimeError{"Parse JSON Error", "error occurred when parsing JSON format response"}, nil
	}
	return nil, loginMsgJson
}

func emitLog(err error, loginMsgJson map[string]interface{}) {
	if err != nil {
		Error.Println(err)
	} else if loginMsgJson["status"] == 1.0 {
		Info.Printf("%v, login user: %v, login ip: %v, login loc: %v\n",
			loginMsgJson["info"],
			loginMsgJson["logout_username"],
			loginMsgJson["logout_ip"],
			loginMsgJson["logout_location"])
	} else {
		Info.Println(loginMsgJson["info"])
	}
}

func runInLoop(param url.Values, interval int) error {
	for {
		err, loginMsgJson := loginRequest(param, interval)
		emitLog(err, loginMsgJson)
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return nil
}

func runOnce(param url.Values) error {
	err, loginMsgJson := loginRequest(param, 0)
	emitLog(err, loginMsgJson)
	return nil
}

func checkOptions(options *Options) error {
	if options.username == "" || options.password == "" {
		return &RuntimeError{"Command Parse Error", "username and password are required."}
	} else if options.interval < 0 {
		return &RuntimeError{"Command Parse Error", "-i option cannot be less than 0."}
	}
	return nil
}
