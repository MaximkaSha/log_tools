package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	Server        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`                    // 0 for sync
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"` // empty for no store
	RestoreFlag   bool          `env:"RESTORE" envDefault:"true"`                           //restore from file
}

type Server struct {
	cfg   Config
	handl handlers.Handlers
}

func NewServer() Server {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	repo := storage.NewRepo()
	handl := handlers.NewHandlers(repo)
	return Server{
		cfg:   cfg,
		handl: handl,
	}
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

	fmt.Println("Server is listening...")
	log.Fatal(http.ListenAndServe(s.cfg.Server, mux))
	s.saveData(s.cfg.StoreFile)
}

func (s *Server) routins(cfg *Config) {
	log.Println("start routiner.")
	//sigc := make(chan os.Signal, 1)
	//signal.Notify(sigc,
	//	syscall.SIGINT,
	//	syscall.SIGTERM,
	//	syscall.SIGQUIT)
	tickerStore := time.NewTicker(cfg.StoreInterval)
	defer tickerStore.Stop()
	for {
		select {
		case <-tickerStore.C:
			s.saveData(cfg.StoreFile)
			//	case <-sigc:
			//		s.saveData(cfg.StoreFile)
			//		log.Println("Exiting")
			//		return
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
