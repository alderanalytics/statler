package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/alderanalytics/statler/client"
	"github.com/codegangsta/cli"
	"github.com/stathat/go"
	"github.com/ugorji/go/codec"
)

var (
	cfgBindAddress     string
	cfgStathatAPIKey   string
	cfgStatHatReporter *stathat.Reporter
)

var (
	errUnknownStatKind = errors.New("unknown stat kind")
)

func statReport(stat *statler.Stat) error {
	switch stat.Kind {
	case statler.Value:
		return cfgStatHatReporter.PostEZValue(stat.Key, cfgStathatAPIKey, stat.Value)
	case statler.Count:
		return cfgStatHatReporter.PostEZCount(stat.Key, cfgStathatAPIKey, int(stat.Count))
	}

	return errUnknownStatKind
}

func statHandler(ln *net.UDPConn) {
	buf := make([]byte, 1024)
	var h codec.Handle = new(codec.MsgpackHandle)

	var stat statler.Stat
	for {
		n, _, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		err = codec.NewDecoderBytes(buf[:n], h).Decode(&stat)
		if err != nil {
			log.Println(err)
			continue
		}

		err = statReport(&stat)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func cmdServe(ctx *cli.Context) {
	addr, err := net.ResolveUDPAddr("udp", cfgBindAddress)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()
	log.Println("listening on", cfgBindAddress)

	quit := make(chan bool)
	for i := 0; i < runtime.NumCPU(); i++ {
		go statHandler(ln)
	}
	<-quit
}

func main() {
	app := cli.NewApp()
	app.Name = "statler"
	app.Usage = "collect and relay stats"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "bindAddress",
			Value:       ":5354",
			Usage:       "bind address for metrics collection",
			EnvVar:      "STATLER_BIND_ADDRESS",
			Destination: &cfgBindAddress,
		},
		cli.StringFlag{
			Name:        "stathatAPIKey",
			Usage:       "stathat api key for relay",
			EnvVar:      "STATLER_STATHAT_API_KEY",
			Destination: &cfgStathatAPIKey,
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "serve",
			Usage:  "start the server",
			Action: cmdServe,
		},
	}

	app.Before = func(*cli.Context) error {
		tr := &http.Transport{MaxIdleConnsPerHost: 40}
		cfgStatHatReporter = stathat.NewReporter(100000, 40, tr)
		return nil
	}

	app.Run(os.Args)
}
