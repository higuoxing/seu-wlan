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
func (w *Worker) Login(ctx context.Context, form url.Values, infoch chan<- string, errch chan<- error) {
	req, err := http.NewRequest("POST", "https://w.seu.edu.cn/index.php/index/login", strings.NewReader(form.Encode()))
	if err != nil {
		errch <- fmt.Errorf("HTTP Request Error: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	resp, err := w.Client.Do(req)
	if err != nil {
		select {
		// If this error is caused by canceling context, then we suppress this error.
		case <-ctx.Done():
			return
		default:
			errch <- err
			return
		}
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
		infoch <- fmt.Sprintf("%v login user: %v login ip: %v login loc: %v\n",
			loginMsgJSON["info"],
			loginMsgJSON["logout_username"],
			loginMsgJSON["logout_ip"],
			loginMsgJSON["logout_location"])
		return
	}

	infoch <- fmt.Sprintf("%v", loginMsgJSON["info"])

	return
}
