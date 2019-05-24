package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/vgxbj/seu-wlan/pkg/config"
	"github.com/vgxbj/seu-wlan/pkg/logger"
	"github.com/vgxbj/seu-wlan/pkg/worker"
)

var (
	options *config.Options
	log     *logger.Logger
)

func init() {
	options = &config.Options{}
	log = logger.NewLogger()

	flag.StringVar(&options.Username, "u", "", "Your card number. (Required)")
	flag.StringVar(&options.Password, "p", "", "Your password. (Required)")
	flag.StringVar(&options.ConfigFile, "c", "", "Your config file.")
	flag.IntVar(&options.Interval, "i", 0, "Run this tool periodically.")
	flag.IntVar(&options.Timeout, "timeout", 1, "Timeout of login request.")
	flag.IntVar(&options.Workers, "workers", 1, "Number of workers to send request. (Experimental)")
	flag.BoolVar(&options.EnableMacAuth, "enable-mac-auth", false, "Enable this machine's mac address to be remembered.")
	flag.BoolVar(&options.DisableTLSVerification, "disable-tls-verification", false, "Disable TLS certificate verification.")

	flag.Usage = func() {
		fmt.Println("Usage: seu-wlan [options] param")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	err := config.VerifyOptions(options)
	if err != nil {
		log.Errorf("%v", err)
		flag.Usage()
		os.Exit(1)
	}

	workers := worker.Workers(options)
	form := config.EncodePOSTForm(options)
	infoch := make(chan string)
	errch := make(chan error)
	var wg sync.WaitGroup

	for {
		wg.Add(1)

		go func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for _, w := range workers {
				go w.Login(ctx, form, infoch, errch)
			}

			select {
			case info := <-infoch:
				log.Infof("%s", info)
			case err := <-errch:
				log.Errorf("%v", err)
			case <-time.After(time.Duration(options.Timeout) * time.Second):
				log.Errorf("HTTP Request Error: Timeout")
			}

			wg.Done()
		}()

		wg.Wait()
		if options.Interval > 0 {
			<-time.After(time.Duration(options.Interval) * time.Second)
		} else {
			break
		}
	}
}
