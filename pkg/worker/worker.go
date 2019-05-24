package worker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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

// Workers ... Initialize more than one workers.
func Workers(o *config.Options) []*Worker {
	workers := make([]*Worker, o.Workers)

	for i := 0; i < len(workers); i++ {
		workers[i] = NewWorker(o)
	}

	return workers
}

// Login ... Send POST request.
func (w *Worker) Login(ctx context.Context, form url.Values, msg chan<- string, errch chan<- error) {
	// resp, err := w.Client.PostForm("https://w.seu.edu.cn/index.php/index/login", form)
	req, err := http.NewRequest("POST", "https://w.seu.edu.cn/index.php/index/login", strings.NewReader(form.Encode()))
	if err != nil {
		errch <- fmt.Errorf("HTTP Request Error: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	resp, err := w.Client.Do(req)

	if err != nil {
		errch <- fmt.Errorf("HTTP Request Error: %s", "error occurred when sending post request")
		return
	}
	defer resp.Body.Close()

	loginMsgRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errch <- fmt.Errorf("Read Response Error: %s", "error occurred when reading response from server")
		return
	}

	var loginMsgJSON map[string]interface{}
	err = json.Unmarshal(loginMsgRaw, &loginMsgJSON)
	if err != nil {
		errch <- fmt.Errorf("Parsing JSON Error: %s", "error occurred when parsing JSON format response")
		return
	}

	if loginMsgJSON["status"] == 1.0 {
		msg <- fmt.Sprintf("%v login user: %v login ip: %v login loc: %v\n",
			loginMsgJSON["info"],
			loginMsgJSON["logout_username"],
			loginMsgJSON["logout_ip"],
			loginMsgJSON["logout_location"])
		return
	}

	msg <- fmt.Sprintf("%v", loginMsgJSON["info"])

	return
}
