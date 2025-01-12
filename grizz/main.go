package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mosajjal/x/grizz/ip"
	"github.com/phuslu/log"
)

var nocolorLog = strings.ToLower(os.Getenv("NO_COLOR")) == "true"

var (
	cfgFile          = flag.String("config", "", "path to the config file")
	printDefaultConf = flag.Bool("defaultconfig", false, "write the default config to stdout")
)

func serveEndpoint(ep *EndpointCfg) error {
	log.Info().Msgf("Starting endpoint %s", ep.Name)
	log.Info().Msgf("Modes: %v", ep.Modes)
	log.Debug().Msgf("Fetching file %s", ep.File)
	r, err := fetchFile(ep.File)
	if err != nil {
		return fmt.Errorf("failed to fetch file %s: %s", ep.File, err)
	}
	defer r.Close()
	log.Debug().Msgf("Parsing IPNets")
	looker, err := populateLooker(r)
	if err != nil {
		return fmt.Errorf("failed to populate looker: %s", err)
	}

	// start the listeners
	// create an error channel as big as the number of modes
	// this will be used to wait for all the listeners to stop
	errc := make(chan error, len(ep.Modes))
	for _, m := range ep.Modes {
		switch m {
		case "http":
			// create a new HTTP server
			ipsrv := ip.NewFibreServer(looker, ep.HTTPListener, ep.HTTPBasePath)
			// run the server in a goroutine and send the error to the channel
			go func() {
				errc <- ipsrv.ListenAndServe()
			}()
		case "socket":
			// create a new network server
			ipsrv := ip.NewNetworkServer(looker, ep.SocketListener)
			// run the server in a goroutine and send the error to the channel
			go func() {
				errc <- ipsrv.ListenAndServe()
			}()
		}
	}

	// reload the looker every auto_reload minutes
	if ep.AutoReload > 0 {
		go func() error {
			time.Sleep(time.Duration(ep.AutoReload) * time.Minute)
			r, err := fetchFile(ep.File)
			if err != nil {
				return fmt.Errorf("failed to fetch file %s: %s", ep.File, err)
			}
			defer r.Close()
			ipnet, err := getIPNets(r)
			if err != nil {
				return fmt.Errorf("failed to get IPNets: %s", err)
			}

			looker.Reload(ipnet...)
			return nil
		}()
	}

	// wait for all the listeners to stop
	for i := 0; i < len(ep.Modes); i++ {
		if err := <-errc; err != nil {
			return err
		}
	}
	return nil
}

func main() {

	flag.Parse()

	var config Config
	if err := hclsimple.DecodeFile(*cfgFile, nil, &config); err != nil {
		log.Fatal().Msgf("Failed to load configuration: %s", err)
	}

	// validate the configuration
	if err := config.Validate(); err != nil {
		log.Fatal().Msgf("Invalid configuration: %s", err)
	}

	if log.IsTerminal(os.Stderr.Fd()) {
		log.DefaultLogger = log.Logger{
			TimeFormat: "15:04:05",
			Caller:     1,
			Level:      log.ParseLevel(config.LogLevel),
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				QuoteString:    true,
				EndWithMessage: true,
			},
		}
	}
	// Create an error channel the size of the number of endpoints
	// This will be used to wait for all the endpoints to stop
	errc := make(chan error, len(config.Endpoints))

	// Start all the endpoints
	for _, ep := range config.Endpoints {
		go func(ep EndpointCfg) {
			errc <- serveEndpoint(&ep)
		}(ep)
	}

	// Wait for all the endpoints to stop
	for i := 0; i < len(config.Endpoints); i++ {
		if err := <-errc; err != nil {
			log.Error().Msgf("Error: %s", err)
		}
	}

}
