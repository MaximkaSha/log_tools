package server

import (
	"compress/flate"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaximkaSha/log_tools/internal/crypto"
	"github.com/MaximkaSha/log_tools/internal/database"
	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	Server        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`                    // 0 for sync
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"` // empty for no store test.json /tmp/devops-metrics-db.json
	RestoreFlag   bool          `env:"RESTORE" envDefault:"true"`                           //restore from file
	KeyFileFlag   string        `env:"KEY"`                                                 // key
	DatabaseEnv   string        `env:"DATABASE_DSN"`
}

type Server struct {
	cfg   Config
	handl handlers.Handlers
	srv   http.Server
	db    *database.Database
}

func NewServer() Server {
	var cfg Config
	var envCfg = make(map[string]bool)
	opts := env.Options{
		OnSet: func(tag string, value interface{}, isDefault bool) {
			envCfg[tag] = isDefault
		},
	}
	err := env.Parse(&cfg, opts)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()
	var a = flag.Lookup("a")
	if envCfg["ADDRESS"] && a != nil {
		cfg.Server = *srvAdressArg
	}
	a = flag.Lookup("i")
	if envCfg["STORE_INTERVAL"] && a != nil {
		//	log.Printf(storeIntervalArg.String())
		cfg.StoreInterval = *storeIntervalArg
	}
	b := flag.Lookup("d")
	//log.Println(b)
	//log.Println(envCfg["DATABASE_DSN"])
	if !envCfg["DATABASE_DSN"] && b != nil {
		cfg.DatabaseEnv = *databaseArg
	}
	//log.Println(cfg.DatabaseEnv)
	a = flag.Lookup("f")
	if envCfg["STORE_FILE"] && a != nil && b == nil {
		cfg.StoreFile = *storeFileArg
	}
	a = flag.Lookup("r")
	if envCfg["RESTORE"] && a != nil {
		cfg.RestoreFlag = *restoreFlagArg
	}
	a = flag.Lookup("k")
	if envCfg["KEY"] && a != nil {
		cfg.KeyFileFlag = *keyFileArg
	}

	repo := storage.NewRepo()
	cryptoService := crypto.NewCryptoService()
	cryptoService.InitCryptoService(cfg.KeyFileFlag)
	handl := handlers.NewHandlers(repo, cryptoService)
	log.Println(cfg.DatabaseEnv)
	if cfg.DatabaseEnv != "" {
		DB := database.NewDatabase(cfg.DatabaseEnv)
		DB.InitDatabase()
		err := DB.Ping()
		log.Println(err)

		handl.DB = &DB
		return Server{
			cfg:   cfg,
			handl: handl,
			srv:   http.Server{},
			db:    &DB,
		}
	}

	return Server{
		cfg:   cfg,
		handl: handl,
		srv:   http.Server{},
	}
}

var (
	srvAdressArg     *string
	storeIntervalArg *time.Duration
	storeFileArg     *string
	restoreFlagArg   *bool
	keyFileArg       *string
	databaseArg      *string
)

func init() {
	srvAdressArg = flag.String("a", "localhost:8080", "host:port (default localhost:8080)")
	storeIntervalArg = flag.Duration("i", time.Duration(300*time.Second), "store interval in seconds (default 300s)")
	storeFileArg = flag.String("f", "/tmp/devops-metrics-db.json", "path to file for store (default '/tmp/devops-metrics-db.json')")
	restoreFlagArg = flag.Bool("r", true, "if is true restore data from env:RESTORE (default true)")
	keyFileArg = flag.String("k", "", "hmac key")
	databaseArg = flag.String("d", "", "string database config")
}

func (s *Server) StartServe() {
	if s.cfg.DatabaseEnv == "" {
		if s.cfg.StoreFile != "" || s.cfg.StoreInterval.Nanoseconds() > 0 {
			go s.routins(&s.cfg)
		}
	}
	if s.cfg.RestoreFlag {
		s.Restore(s.cfg.StoreFile)
	}
	if s.cfg.StoreInterval == 0 {
		s.handl.SyncFile = s.cfg.StoreFile
	}

	mux := chi.NewRouter()
	compressor := middleware.NewCompressor(flate.DefaultCompression)
	mux.Use(compressor.Handler)
	mux.Post("/update/{type}/{name}/{value}", s.handl.HandleUpdate)
	mux.Get("/value/{type}/{name}", s.handl.HandleGetUpdate)
	mux.Get("/", s.handl.HandleGetHome)
	mux.Get("/ping", s.handl.HandleGetPing)
	mux.Post("/update/", s.handl.HandlePostJSONUpdate)
	mux.Post("/value/", s.handl.HandlePostJSONValue)
	s.srv.Addr = s.cfg.Server
	s.srv.Handler = mux
	fmt.Println("Server is listening...")
	if err := s.srv.ListenAndServe(); err != nil {
		log.Printf("Server shutdown: %s", err.Error())
		if s.cfg.DatabaseEnv != "" {
			s.db.DB.Close()
		}
		s.saveData(s.cfg.StoreFile)
	}
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
				s.db.DB.Close()
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
