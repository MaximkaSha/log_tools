package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	Server        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`                    // 0 for sync
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"` // empty for no store test.json /tmp/devops-metrics-db.json
	RestoreFlag   bool          `env:"RESTORE" envDefault:"true"`                           //restore from file
}

type Server struct {
	cfg   Config
	handl handlers.Handlers
	srv   http.Server
}

func NewServer() Server {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()
	if _, err := os.LookupEnv("ADDRESS"); err {
		cfg.Server = *srvAdressArg
	}
	if _, err := os.LookupEnv("STORE_INTERVAL"); err {
		cfg.StoreInterval = time.Duration(*storeIntervalArg)
	}
	if _, err := os.LookupEnv("POLL_INTERVAL"); err {
		cfg.StoreFile = *storeFileArg
	}
	if _, err := os.LookupEnv("RESTORE"); err {
		cfg.RestoreFlag = *restoreFlagArg
	}

	repo := storage.NewRepo()
	handl := handlers.NewHandlers(repo)
	return Server{
		cfg:   cfg,
		handl: handl,
		srv:   http.Server{},
	}
}

var (
	srvAdressArg     *string
	storeIntervalArg *int
	storeFileArg     *string
	restoreFlagArg   *bool
)

func init() {
	srvAdressArg = flag.String("a", "localhost:8080", "host:port (default localhost:8080)")
	storeIntervalArg = flag.Int("i", 300, "store interval in seconds (default 300s)")
	storeFileArg = flag.String("f", "/tmp/devops-metrics-db.json", "path to file for store (default '/tmp/devops-metrics-db.json')")
	restoreFlagArg = flag.Bool("r", true, "if is true restore data from env:STORE_FILE (default true)")
}

func (s *Server) StartServe() {
	if s.cfg.StoreFile != "" || s.cfg.StoreInterval > 0 {
		go s.routins(&s.cfg)
	}
	if s.cfg.RestoreFlag {
		s.Restore(s.cfg.StoreFile)
	}
	if s.cfg.StoreInterval == 0 {
		s.handl.SyncFile = s.cfg.StoreFile
	}

	mux := chi.NewRouter()
	mux.Post("/update/{type}/{name}/{value}", s.handl.HandleUpdate)
	mux.Get("/value/{type}/{name}", s.handl.HandleGetUpdate)
	mux.Get("/", s.handl.HandleGetHome)
	mux.Post("/update/", s.handl.HandlePostJSONUpdate)
	mux.Post("/value/", s.handl.HandlePostJSONValue)

	s.srv.Addr = s.cfg.Server
	s.srv.Handler = mux
	fmt.Println("Server is listening...")
	log.Fatal(s.srv.ListenAndServe())
	s.saveData(s.cfg.StoreFile)
}

func (s *Server) routins(cfg *Config) {
	log.Println("start routiner.")
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	tickerStore := time.NewTicker(cfg.StoreInterval)
	defer tickerStore.Stop()
	for {
		select {
		case <-tickerStore.C:
			s.saveData(cfg.StoreFile)
		case <-sigc:

			s.saveData(cfg.StoreFile)
			if err := s.srv.Shutdown(context.Background()); err != nil {
				log.Printf("Gracefully Shutdown: %v", err)
			}
			return
		}
	}

}

func (s *Server) saveData(file string) {
	s.handl.Repo.SaveData(file)
	log.Println("Data stored")
}

func (s *Server) Restore(file string) {
	s.handl.Repo.Restore(file)
	//	log.Println("Data restored")
}
