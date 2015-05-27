package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	cgcli "github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/gregjones/httpcache"
	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/gregjones/httpcache/diskcache"
	"github.com/registrobr/rdap-client/bootstrap"
)

func main() {
	cgcli.AppHelpTemplate = `
NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} {{if .Flags}}[global options]{{end}} OBJECT

VERSION:
   {{.Version}}{{if len .Authors}}

AUTHOR(S): 
   {{range .Authors}}{{ . }}{{end}}{{end}}

GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`

	app := cgcli.NewApp()
	app.Name = "rdap"
	app.Usage = "RDAP cgclient"
	app.Author = "NIC.br"

	app.Flags = []cgcli.Flag{
		cgcli.StringFlag{
			Name:  "cache",
			Value: path.Join(os.Getenv("HOME"), ".rdap"),
			Usage: "directory for caching bootstrap and RDAP data",
		},
		cgcli.StringFlag{
			Name:  "bootstrap",
			Value: bootstrap.RDAPBootstrap,
			Usage: "RDAP bootstrap service URL",
		},
		cgcli.BoolFlag{
			Name:  "no-cache",
			Usage: "Don't cache anything",
		},
		cgcli.BoolFlag{
			Name:  "skip-tls-verification,S",
			Usage: "Skip TLS verification",
		},
		cgcli.StringFlag{
			Name:  "host,H",
			Value: "",
			Usage: "Host where to send the query (bypass bootstrap)",
		},
	}

	app.Commands = []cgcli.Command{}
	app.Action = action

	app.Run(os.Args)
}

func action(ctx *cgcli.Context) {
	var (
		cache               = ctx.String("cache")
		bootstrapURI        = ctx.String("bootstrap")
		host                = ctx.String("host")
		skipTLSVerification = ctx.Bool("skip-tls-verification")
		httpClient          = &http.Client{}
		bs                  *bootstrap.Client
		uris                []string
	)

	if !ctx.Bool("no-cache") {
		transport := httpcache.NewTransport(
			diskcache.New(cache),
		)

		transport.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerification},
		}

		httpClient.Transport = transport
	}

	if len(host) == 0 {
		bs = bootstrap.NewClient(httpClient)

		if len(bootstrapURI) > 0 {
			bs.Bootstrap = bootstrapURI
		}
	} else {
		if _, err := url.Parse(host); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		uris = []string{host}
	}

	object := strings.Join(ctx.Args(), " ")
	if object == "" {
		cgcli.ShowAppHelp(ctx)
		os.Exit(1)
	}

	cli := &cli{
		uris:       uris,
		httpClient: httpClient,
		bootstrap:  bs,
		wr:         os.Stdout,
	}

	handlers := []handler{
		cli.asn(),
		cli.ip(),
		cli.ipnetwork(),
		cli.domain(),
		cli.entity(),
	}

	for _, handler := range handlers {
		ok, err := handler(object)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		if ok {
			break
		}
	}

	os.Exit(0)
}
