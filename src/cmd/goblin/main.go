package main

import (
	"flag"
	"fmt"
	"goblin/src/internals/config"
	"goblin/src/internals/controller"
	"goblin/src/internals/logger"
	"goblin/src/internals/policy"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

const (
	DefaultPort   int    = 8080
	DefaultTarget string = "http://localhost:8888"
	ParamUrl      string = "UPSTREAM_URL"
	ParamPort     string = "PORT"
)

type Settings struct {
	Config     Config
	init       bool
	verbose    bool
	policyPath string
}
type Config struct {
	Port        int
	UpstreamURL *url.URL
}
type chaosTransport struct {
	next http.RoundTripper
}

/*
this is where we will inject chaos
*/
func (c *chaosTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.next.RoundTrip(req)
}

func LoadSettings() (*Settings, error) {
	var config Config

	upstreamURL := flag.String("url", DefaultTarget, "Upstream URL")
	port := flag.Int("port", DefaultPort, "Listen port")
	init := flag.Bool("init", false, "Save config to `goblin.toml`")

	flag.Parse()

	target, err := url.Parse(*upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", *upstreamURL, err)
	}

	if target.Scheme != "http" && target.Scheme != "https" {
		return nil, fmt.Errorf("upstream URL must use http or https")
	}
	if target.Host == "" {
		return nil, fmt.Errorf("upstream URL must include a host")
	}
	if *port < 1 || *port > 65535 {
		return nil, fmt.Errorf("invalid port %d", *port)
	}

	config.UpstreamURL = target
	config.Port = *port

	log.Printf("Upstream [%s]", config.UpstreamURL)
	settings := Settings{
		Config: config,
		init:   *init,
	}

	return &settings, nil
}

func main() {
	logger := logger.NewJSONLogger()
	settings, err := LoadSettings()
	if err != nil {
		log.Fatalf("Invalid settings: %v", err)
		os.Exit(1)
	}

	// initial run create file with defaults
	if settings.init {
		store := config.NewConfigStore(config.ChaosConfig{})
		store.Save("goblin.toml")
		log.Print("Saving configuration file")
		return
	}

	// read policy from disk
	store, err := config.LoadFromFile()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
		os.Exit(1)
	}
	policies := policy.LoadPolicies(store, logger)

	proxy := httputil.NewSingleHostReverseProxy(settings.Config.UpstreamURL)

	transport := policy.ChainTransport(
		http.DefaultTransport,
		policies.Transport...,
	)

	// notes for other ideas
	// transport := policy.ChainTransport(
	// 	http.DefaultTransport,
	// 	policy.DelayPolicy(store),
	// 	policy.FailurePolicy(store),
	// 	//RequestHeaderPolicy(store),
	// )

	proxy.Transport = transport

	proxy.ModifyResponse = policy.ChainResponse(
		policies.Response...,
	// notes for other ideas
	// policy.DelayResponsePolicy(store),
	// policy.StatusPolicy(store),
	// policy.HeaderPolicy(store),
	//ResponseBodyPolicy(store),
	)

	renderer, err := controller.NewRenderer(logger)
	if err != nil {
		log.Fatalf("could not load templates: %v", err)
		os.Exit(1)
	}
	webhandler := controller.New(store, renderer, logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/gablin") {
			webhandler.Http(w, r)
			return
		}
		/*
			block local items within admin
			since I'm using this currently just for Backend testing
			this is fine but if I switch to FE at some point I
			will want to revisit this idea.  To be fair this is just
			to solve an annoyance when I'm loading the admin panel
		*/
		if controller.MatchNoopPaths(r.URL.Path) {
			w.WriteHeader(http.StatusOK)
			return
		}
		logger.Info("Url", "path", r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	log.Printf("Goblin is running on port :%d", settings.Config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", settings.Config.Port), handler))

}
