package worker

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/vgxbj/seu-wlan/pkg/config"
)

// Worker ... Wrapper for http.Client.
type Worker struct {
	Client *http.Client
}

// NewWorker ... New worker.
func NewWorker(o *config.Options) *Worker {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: o.DisableTLSVerification,
		},
	}

	c := &http.Client{
		Transport: t,
		Timeout:   time.Duration(o.Timeout) * time.Second,
	}

	return &Worker{c}
}

// Login ... Send POST request.
func (w *Worker) Login(url string, form url.Values) (string, error) {
	resp, err := w.Client.PostForm(url, form)
	if err != nil {
		return "", fmt.Errorf("HTTP Request Error: %s", "error occurred when sending post request")
	}
	defer resp.Body.Close()

	loginMsgRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Read Response Error: %s", "error occurred when reading response from server")
	}

	var loginMsgJSON map[string]interface{}
	err = json.Unmarshal(loginMsgRaw, &loginMsgJSON)
	if err != nil {
		return "", fmt.Errorf("Parsing JSON Error: %s", "error occurred when parsing JSON format response")
	}

	if loginMsgJSON["status"] == 1.0 {
		return fmt.Sprintf("%v\tlogin user: %v\tlogin ip: %v\tlogin loc: %v\n",
			loginMsgJSON["info"],
			loginMsgJSON["logout_username"],
			loginMsgJSON["logout_ip"],
			loginMsgJSON["logout_location"]), nil
	}

	return fmt.Sprintf("%v", loginMsgJSON["info"]), nil
}
