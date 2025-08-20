package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/complytime/complybeacon/compass/cmd/compass/server"
	compass "github.com/complytime/complybeacon/compass/service"
)

func main() {

	var (
		port, catalogPath, configPath string
		skipTLS                       bool
	)

	flag.StringVar(&port, "port", "8080", "Port for HTTP server")
	flag.BoolVar(&skipTLS, "skip-tls", false, "Run without TLS")

	// TODO: This needs to be come Layer 3 policy and complete resolution on startup
	flag.StringVar(&catalogPath, "catalog", "./hack/sampledata/osps.yaml", "Path to Layer 2 catalog")
	flag.StringVar(&configPath, "config", "./docs/config.yaml", "Path to compass config file")
	flag.Parse()

	catalogPath = filepath.Clean(catalogPath)
	scope, err := server.NewScopeFromCatalogPath(catalogPath)
	if err != nil {
		log.Fatal(err)
	}

	var cfg server.Config
	configPath = filepath.Clean(configPath)
	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	transformers, err := server.NewTransformerSet(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	service := compass.NewService(transformers, scope)

	s := server.NewGinServer(service, port)

	if skipTLS {
		log.Println(`Warning: Insecure connections permitted. TLS is highly recommended for production.`)
		log.Fatal(s.ListenAndServe())
	} else {
		cert, key := server.SetupTLS(s, cfg)
		log.Fatal(s.ListenAndServeTLS(cert, key))
	}
}
