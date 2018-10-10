package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
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

func logger_init() {
	Info = log.New(os.Stdout, "[Info]    ", log.Ldate|log.Ltime)
	Warning = log.New(os.Stdout, "[Warning] ", log.Ldate|log.Ltime)
	Error = log.New(os.Stdout, "[Error]   ", log.Ldate|log.Ltime)
}

func encode_param(username, password string, macauth int) url.Values {
	b64pass := base64.StdEncoding.EncodeToString([]byte(password))
	return url.Values{"username": {username},
		"password":      {string(b64pass)},
		"enablemacauth": {string(macauth)}}
}

func login_request(param url.Values) (error, map[string]interface{}) {
	response, err := http.PostForm(SEU_WLAN_LOGIN_URL, param)
	if err != nil {
		return err, nil
	}
	defer response.Body.Close()

	login_msg_raw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err, nil
	}

	var login_msg_json map[string]interface{}
	err = json.Unmarshal(login_msg_raw, &login_msg_json)
	if err != nil {
		return err, nil
	}
	return nil, login_msg_json
}

func emit_log(err error, login_msg_json map[string]interface{}) {
	if err != nil {
		Error.Printf("network error.\n")
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

func main() {
	// command line options:
	//   user: username
	//   pass: password
	//   macauth: allow seu-wlan server remember your mac address
	username := flag.String("u", "", "Your card number. (Required)")
	password := flag.String("p", "", "Your password. (Required)")
	macauth := flag.Int("m", 0, "Enable seu-wlan remember your mac address. 0 (default) or 1.")
	interval := flag.Int("i", 0, "Enable this plugin run in loop and request seu-wlan login server.")
	flag.Parse()

	if *username == "" || *password == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else if *interval < 0 {
		log.Fatalln("ERROR: interval param -i cannot less than 0.")
	}

	logger_init()

	param := encode_param(*username, *password, *macauth)

	if *interval > 0 {
		err := run_in_loop(param, *interval)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
	} else {
		err := run_once(param)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
	}
}
