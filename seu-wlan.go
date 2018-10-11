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
	err_type string
	err_hint string
}

func (err *RuntimeError) Error() string {
	return fmt.Sprintf("[%v]  %v", err.err_type, err.err_hint)
}

func logger_init() {
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
	logger_init()
}

func main() {
	flag.Parse()

	err := check_options(options)
	if err != nil {
		Error.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	param := encode_param(options)

	if options.interval > 0 {
		err := run_in_loop(param, options.interval)
		if err != nil {
			Error.Println(err)
			os.Exit(1)
		}
	} else {
		err := run_once(param)
		if err != nil {
			Error.Println(err)
			os.Exit(1)
		}
	}
}

func encode_param(options *Options) url.Values {
	b64pass := base64.StdEncoding.EncodeToString([]byte(options.password))
	return url.Values{"username": {options.username},
		"password":      {string(b64pass)},
		"enablemacauth": {string(options.macauth)}}
}

func login_request(param url.Values) (error, map[string]interface{}) {
	response, err := http.PostForm(SEU_WLAN_LOGIN_URL, param)
	if err != nil {
		return &RuntimeError{"HTTP Request Error", "error occured when sending post request"}, nil
	}
	defer response.Body.Close()

	login_msg_raw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &RuntimeError{"Read Response Error", "error occured when reading response from server"}, nil
	}

	var login_msg_json map[string]interface{}
	err = json.Unmarshal(login_msg_raw, &login_msg_json)
	if err != nil {
		return &RuntimeError{"Parse JSON Error", "error occured when parsing JSON format response"}, nil
	}
	return nil, login_msg_json
}

func emit_log(err error, login_msg_json map[string]interface{}) {
	if err != nil {
		Error.Println(err)
	} else if login_msg_json["status"] == 1.0 {
		Info.Printf("%v, login user: %v, login ip: %v, login loc: %v\n",
			login_msg_json["info"],
			login_msg_json["logout_username"],
			login_msg_json["logout_ip"],
			login_msg_json["logout_location"])
	} else {
		Info.Println(login_msg_json["info"])
	}
}

func run_in_loop(param url.Values, interval int) error {
	for {
		err, login_msg_json := login_request(param)
		emit_log(err, login_msg_json)
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return nil
}

func run_once(param url.Values) error {
	err, login_msg_json := login_request(param)
	emit_log(err, login_msg_json)
	return nil
}

func check_options(options *Options) error {
	if options.username == "" || options.password == "" {
		return &RuntimeError{"Command Parse Error", "username and password are required."}
	} else if options.interval < 0 {
		return &RuntimeError{"Command Parse Error", "-i option cannot be less than 0."}
	}
	return nil
}
