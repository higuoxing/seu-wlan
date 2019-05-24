package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/vgxbj/seu-wlan/pkg/config"
	"github.com/vgxbj/seu-wlan/pkg/logger"
	"github.com/vgxbj/seu-wlan/pkg/worker"
)

var (
	o *config.Options
	l *logger.Logger
)

func init() {
	o = &config.Options{}
	l = logger.NewLogger()

	flag.StringVar(&o.Username, "u", "", "Your card number. (Required)")
	flag.StringVar(&o.Password, "p", "", "Your password. (Required)")
	flag.StringVar(&o.ConfigFile, "c", "", "Your config file.")
	flag.IntVar(&o.Interval, "i", 0, "Run this tool periodically.")
	flag.IntVar(&o.Timeout, "timeout", 1, "Timeout of login request.")
	flag.IntVar(&o.Workers, "workers", 1, "Number of workers to send request. (Experimental)")
	flag.BoolVar(&o.EnableMacAuth, "enable-mac-auth", false, "Enable this machine's mac address to be remembered.")
	flag.BoolVar(&o.DisableTLSVerification, "disable-tls-verification", false, "Disable TLS certificate verification.")

	flag.Usage = func() {
		fmt.Println("Usage: seu-wlan [options] param")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	err := config.VerifyOptions(o)
	if err != nil {
		l.Errorf("%v", err)
		flag.Usage()
		os.Exit(1)
	}

	workers := worker.Workers(o)
	form := config.EncodePOSTForm(o)
	infoch := make(chan string)
	errch := make(chan error)

	for {
		go func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for _, w := range workers {
				w.Login(ctx, form, infoch, errch)
			}
		}()

		select {
		case s := <-infoch:
			l.Infof("%s", s)
		case errch := <-errch:
			l.Errorf("%v", errch)
		case <-time.After(time.Duration(o.Timeout) * time.Second):
			l.Errorf("HTTP Request Error: Timeout")
		}

		if o.Interval > 0 {
			<-time.After(time.Duration(o.Interval) * time.Second)
		} else {
			break
		}
	}
}
