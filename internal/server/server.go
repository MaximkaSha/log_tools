//  Server module collects runtime metrics and save recieved from remote agent.
//  Moodule is controlled by enviroment variables and console keys.
//  All settings are provided in console output.
package server

import (
	"compress/flate"
	"context"
	"net"
	"strings"

	//	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaximkaSha/log_tools/internal/crypto"
	"github.com/MaximkaSha/log_tools/internal/database"
	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/models"
	pb "github.com/MaximkaSha/log_tools/internal/proto"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Config structure is server configiguration.
type Config struct {
	Server          string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreFile       string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	KeyFileFlag     string        `env:"KEY" envDefault:"12345678"`
	DatabaseEnv     string        `env:"DATABASE_DSN"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	RestoreFlag     bool          `env:"RESTORE" envDefault:"true"`
	PrivateKeyFile  string        `env:"CRYPTO_KEY"`
	configFile      string        `env:"CONFIG"`
	TrustedSubnet   string        `env:"TRUSTED_SUBNET"`
	CertGRPCFile    string        `env:"CERT_FILE"`
	CertKeyGRPCFile string        `env:"CERT_KEY_FILE"`
}

func (c *Config) isDefault(flagName string, envName string) bool {
	flagPresent := false
	envPresent := false
	if flag := flag.Lookup(flagName); flag != nil && flag.Value.String() != flag.DefValue {
		flagPresent = true
	}
	if _, ok := os.LookupEnv(envName); ok {
		envPresent = true
	}
	return flagPresent || envPresent
}
func (c *Config) UmarshalJSON(data []byte) (err error) {
	var tmp struct {
		Server          string `json:"address" env:"ADDRESS" envDefault:"localhost:8080"`
		StoreFile       string `json:"store_file" env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
		KeyFileFlag     string `env:"KEY" envDefault:"12345678"`
		DatabaseEnv     string `json:"database_dsn" env:"DATABASE_DSN"`
		StoreInterval   string `json:"store_interval" env:"STORE_INTERVAL" envDefault:"300s"`
		RestoreFlag     bool   `json:"restore" env:"RESTORE" envDefault:"true"`
		PrivateKeyFile  string `json:"crypto_key" env:"CRYPTO_KEY"`
		configFile      string `env:"CONFIG"`
		TrustedSubnet   string `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
		CertGRPCFile    string `json:"cert" env:"CERT_FILE" envDefault:""`
		CertKeyGRPCFile string `json:"cert_key" env:"CERT_KEY_FILE" envDefault:""`
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if !c.isDefault("a", "ADDRESS") {
		c.Server = tmp.Server
	}
	if !c.isDefault("f", "STORE_FILE") {
		c.StoreFile = tmp.StoreFile
	}
	if !c.isDefault("d", "DATABASE_DSN") {
		c.DatabaseEnv = tmp.DatabaseEnv
	}
	if !c.isDefault("i", "STORE_INTERVAL") {
		c.StoreInterval, err = time.ParseDuration(tmp.StoreInterval)
	}
	if !c.isDefault("crypto-key", "CRYPTO_KEY") {
		c.PrivateKeyFile = tmp.PrivateKeyFile
	}
	if !c.isDefault("cert", "CERT_FILE") {
		c.CertGRPCFile = tmp.CertGRPCFile
	}
	if !c.isDefault("cert-key", "CERT_KEY_FILE") {
		c.CertKeyGRPCFile = tmp.CertKeyGRPCFile
	}
	if !c.isDefault("r", "RESTORE") {
		c.RestoreFlag = tmp.RestoreFlag
	}
	return err
}

// Server - internal server structure.
type Server struct {
	handl handlers.Handlers
	srv   *http.Server
	db    *database.Database
	cfg   Config
}

// NewServer - Server constructor.
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
		cfg.StoreInterval = *storeIntervalArg
	}
	b := flag.Lookup("d")
	_, present := os.LookupEnv("DATABASE_DSN")
	if !present && b != nil {
		cfg.DatabaseEnv = *databaseArg
	}
	a = flag.Lookup("f")
	if envCfg["STORE_FILE"] && a != nil {
		cfg.StoreFile = *storeFileArg
	}
	a = flag.Lookup("r")
	if envCfg["RESTORE"] && a != nil {
		cfg.RestoreFlag = *restoreFlagArg
	}
	a = flag.Lookup("t")
	if envCfg["TRUSTED_SUBNET"] || a != nil {
		cfg.TrustedSubnet = *trustedSubnet
	}
	a = flag.Lookup("k")
	if envCfg["KEY"] && a != nil {
		cfg.KeyFileFlag = *keyFileArg
	}
	a = flag.Lookup("crypto-key")
	if envCfg["CRYPTO_KEY"] || a != nil {
		cfg.PrivateKeyFile = *PrivateKeyFileArg
	}

	a = flag.Lookup("cert")
	if envCfg["CERT_FILE"] || a != nil {
		cfg.CertGRPCFile = *certFile
	}
	a = flag.Lookup("cert-file")
	if envCfg["CERT_KEY_FILE"] || a != nil {
		cfg.CertKeyGRPCFile = *keyCertFile
	}

	if envCfg["CONFIG"] || a != nil {
		cfg.configFile = *configFile
	} else {
		a = flag.Lookup("config")
		if envCfg["CONFIG"] || a != nil {
			cfg.configFile = *configFile
		}
	}
	if cfg.configFile != "" {
		jsonData, err := ioutil.ReadFile(cfg.configFile)
		if err != nil {
			log.Println(err)
		}
		err = cfg.UmarshalJSON(jsonData)
		if err != nil {
			log.Println(err)
		}
	}
	var serv = Server{}
	serv.cfg = cfg
	var repo models.Storager
	if cfg.DatabaseEnv == "" {
		imMemory := storage.NewRepo()
		repo = &imMemory
	} else {
		DB := database.NewDatabase(cfg.DatabaseEnv)
		repo = &DB
		DB.InitDatabase()
		serv.db = &DB
	}
	cryptoService := crypto.NewCryptoService()
	cryptoService.InitCryptoService(cfg.KeyFileFlag)
	handl := handlers.NewHandlers(repo, cryptoService, *PrivateKeyFileArg)
	serv.handl = handl
	serv.srv = &http.Server{}
	return serv
}

var (
	srvAdressArg      *string
	storeIntervalArg  *time.Duration
	storeFileArg      *string
	restoreFlagArg    *bool
	keyFileArg        *string
	databaseArg       *string
	PrivateKeyFileArg *string
	configFile        *string
	trustedSubnet     *string
	certFile          *string
	keyCertFile       *string
)

func init() {
	srvAdressArg = flag.String("a", "localhost:8080", "host:port (default localhost:8080)")
	storeIntervalArg = flag.Duration("i", time.Duration(300*time.Second), "store interval in seconds (default 300s)")
	storeFileArg = flag.String("f", "/tmp/devops-metrics-db.json", "path to file for store (default '/tmp/devops-metrics-db.json')")
	restoreFlagArg = flag.Bool("r", true, "if is true restore data from env:RESTORE (default true)")
	keyFileArg = flag.String("k", "", "hmac key")
	databaseArg = flag.String("d", "", "string database config")
	PrivateKeyFileArg = flag.String("crypto-key", "", "private key")
	configFile = flag.String("c", "", "json config file path")
	configFile = flag.String("config", "", "json config file path")
	trustedSubnet = flag.String("t", "", "trusted subnet")
	certFile = flag.String("cert", "", "tls cert file path for gRPC")
	keyCertFile = flag.String("cert-key", "", "tls key for cert file path for gRPC")
}

// StartServe - main server func.
// It stands for endpoits initialization and server handling.
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
	if s.cfg.TrustedSubnet != "" {
		_, ipRange, err := net.ParseCIDR(s.cfg.TrustedSubnet)
		if err == nil {
			log.Printf("trusted subnet set: %s ", ipRange.String())
			s.handl.TrustedSubnet = ipRange
			mux.Use(s.handl.CheckIPMiddleWare)
		} else {
			log.Printf("Error parsing CIDR: %s", err)
			os.Exit(1)
		}

	}
	mux.Post("/update/{type}/{name}/{value}", s.handl.HandleUpdate)
	mux.Get("/value/{type}/{name}", s.handl.HandleGetUpdate)
	mux.Get("/", s.handl.HandleGetHome)
	mux.Get("/ping", s.handl.HandleGetPing)
	mux.Post("/update/", s.handl.HandlePostJSONUpdate)
	mux.Post("/updates/", s.handl.HandlePostJSONUpdates)
	mux.Post("/value/", s.handl.HandlePostJSONValue)
	s.srv.Addr = s.cfg.Server
	s.srv.Handler = mux

	// gPRC server starts.
	fqdn := strings.Split(s.cfg.Server, ":")
	listen, err := net.Listen("tcp", fqdn[0]+":3200")
	if err != nil {
		log.Fatal(err)
	}
	gsrv := grpc.NewServer()
	creds := insecure.NewCredentials()
	if s.cfg.CertGRPCFile != "" {
		creds, err = credentials.NewServerTLSFromFile("cert.pem", "key.pem")
		if err != nil {
			log.Println(err)
		}
		gsrv = grpc.NewServer(grpc.Creds(creds))
	}
	if s.cfg.TrustedSubnet != "" {
		gsrv = grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(s.unaryCheckIPInterceptor))
	}
	pb.RegisterMetricsServer(gsrv, &MetricsServer{
		handl: s.handl,
	})

	fmt.Println("Сервер gRPC начал работу")
	go gsrv.Serve(listen)

	// HTTP server starts
	fmt.Println("Server is listening...")
	if err := s.srv.ListenAndServe(); err != nil {
		log.Printf("Server shutdown: %s", err.Error())
		if s.db != nil {
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
	// 	log.Println("Data restored")
}

type MetricsServer struct {
	pb.UnimplementedMetricsServer

	handl handlers.Handlers
}

func (m MetricsServer) AddMetric(ctx context.Context, in *pb.AddMetricRequest) (*pb.AddMetricResponse, error) {
	log.Println(in.Metric)
	data := models.NewMetric(
		in.Metric.Id,
		strings.ToLower(in.Metric.Mtype.String()),
		&in.GetMetric().Delta,
		&in.GetMetric().Value,
		in.Metric.Hash)
	log.Println(data)
	if m.handl.CryptoService.IsEnable {
		if m.handl.CryptoService.CheckHash(data) {
			response := pb.AddMetricResponse{
				Error: "hash error",
			}
			return &response, status.Errorf(codes.DataLoss, `Check sign error`)
		}
	}
	err := m.handl.Repo.InsertMetric(ctx, data)
	if err != nil {
		response := pb.AddMetricResponse{
			Error: "Bad data",
		}
		return &response, nil
	}
	response := pb.AddMetricResponse{
		Error: "",
	}
	log.Println("END")
	return &response, nil
}

func (m MetricsServer) AddMetrics(ctx context.Context, in *pb.AddMetricsRequest) (*pb.AddMetricsResponse, error) {
	allData := []models.Metrics{}
	// Вот тут не очень конечно. Конвертируем из json в protobuf, потом опять из protobuf в json.
	// Надо наверное сделать две раздельные хранилки в агенте для json и protobuff.

	for i := range in.Metrics {
		data := models.NewMetric(
			in.Metrics[i].Id,
			strings.ToLower(in.Metrics[i].Mtype.String()),
			&in.Metrics[i].Delta,
			&in.Metrics[i].Value,
			in.Metrics[i].Hash)
		// кроме того я так и не понял как получить доступ к данным из интерцептора
		if m.handl.CryptoService.IsEnable {
			if m.handl.CryptoService.CheckHash(data) {
				response := pb.AddMetricsResponse{
					Error: "hash error",
				}
				return &response, status.Errorf(codes.DataLoss, `Check sign error`)
			}
		}
		allData = append(allData, data)
	}
	m.handl.Repo.BatchInsert(ctx, allData)
	response := pb.AddMetricsResponse{
		Error: "",
	}
	return &response, nil
}

func (s Server) unaryCheckIPInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
	}
	data, ok := md["x-real-ip"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing IP")
	}
	ip := net.ParseIP(data[0])
	if ip == nil || !s.handl.TrustedSubnet.Contains(ip) {
		return nil, status.Errorf(codes.Unauthenticated, "not in trusted subnet")
	}

	h, err := handler(ctx, req)
	return h, err

}
