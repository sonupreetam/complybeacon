package push

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/complytime/complybeacon/proofwatch/evidence"
)

// Config listens for evidence push messages
type Config struct {
	ListenAddress string `yaml:"listenAddress"`
}

type Source struct {
	config  *Config
	server  *http.Server
	emitter evidence.InstrumentationFn
}

func NewSource(config *Config, emitter evidence.InstrumentationFn) *Source {
	return &Source{
		config:  config,
		emitter: emitter,
	}
}

func (s *Source) Name() string {
	return "push"
}

func (s *Source) Run(ctx context.Context) error {
	log.Printf("%s running\n", s.Name())
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/push", s.push)

	srv := &http.Server{
		Addr:              s.config.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	s.server = srv

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	return nil
}

func (s *Source) push(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	rawEvidence := evidence.RawEvidence{}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &rawEvidence)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Processing evidence id %s", rawEvidence.ID)

	err = s.emitter(context.Background(), rawEvidence)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Evidence processing completed for %s", rawEvidence.ID)
	w.WriteHeader(http.StatusNoContent)
}

// Stop the target.
func (s *Source) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
