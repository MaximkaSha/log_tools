package agent

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/MaximkaSha/log_tools/internal/ciphers"
	"github.com/MaximkaSha/log_tools/internal/crypto"
	"github.com/MaximkaSha/log_tools/internal/models"
	pb "github.com/MaximkaSha/log_tools/internal/proto"
	"github.com/MaximkaSha/log_tools/internal/utils"
	"github.com/caarlos0/env/v6"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Config structure is agent configiguration.
type Config struct {
	Server         string        `json:"address" env:"ADDRESS" envDefault:"localhost:8080"`
	KeyFile        string        `env:"KEY" envDefault:"key.txt"`
	ReportInterval time.Duration `json:"report_interval" env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `json:"poll_interval" env:"POLL_INTERVAL,required" envDefault:"2s"`
	PublicKeyFile  string        `json:"crypto_key" env:"CRYPTO_KEY" envDefault:""`
	CertGRPCFile   string        `json:"cert" env:"CERT_FILE" envDefault:""`
	configFile     string        `env:"CONFIG"`
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
		Server         string `json:"address" env:"ADDRESS" envDefault:"localhost:8080"`
		KeyFile        string `env:"KEY" envDefault:"key.txt"`
		ReportInterval string `json:"report_interval" env:"REPORT_INTERVAL" envDefault:"10s"`
		PollInterval   string `json:"poll_interval" env:"POLL_INTERVAL,required" envDefault:"2s"`
		PublicKeyFile  string `json:"crypto_key" env:"CRYPTO_KEY" envDefault:""`
		CertGRPCFile   string `json:"cert" env:"CERT_FILE" envDefault:""`
	}
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if !c.isDefault("a", "ADDRESS") {
		c.Server = tmp.Server
	}
	if !c.isDefault("r", "REPORT_INTERVAL") {
		c.ReportInterval, err = time.ParseDuration(tmp.ReportInterval)
		if err != nil {
			return err
		}
	}
	if !c.isDefault("p", "POLL_INTERVAL") {
		c.PollInterval, err = time.ParseDuration(tmp.PollInterval)
	}
	if !c.isDefault("crypto-key", "CRYPTO_KEY") {
		c.PublicKeyFile = tmp.PublicKeyFile
	}
	if !c.isDefault("cert", "CERT_FILE") {
		c.CertGRPCFile = tmp.CertGRPCFile
	}
	return err
}

// Agent collects runtime metrics. Main module of agent.
type Agent struct {
	logDB   []models.Metrics
	cfg     Config
	counter int64
	pubKey  *rsa.PublicKey
	IP      string
}

// NewAgent - Agent constructor.
func NewAgent() Agent {
	ip, err := utils.ExternalIP()
	if err != nil {
		ip = ""
		log.Println("cant get Ip!")
	}
	return Agent{
		logDB:   []models.Metrics{},
		counter: 0,
		cfg:     parseCfg(),
		IP:      ip,
	}
}

// AppendMetric - add given models.Metrics to storage.
func (a *Agent) AppendMetric(m models.Metrics) {
	for i := range a.logDB {
		if a.logDB[i].ID == m.ID {
			a.logDB[i].Delta = m.Delta
			a.logDB[i].Value = m.Value
			return
		}
	}
	a.logDB = append(a.logDB, m)
}

// StartService - main function.
func (a *Agent) StartService() {
	var pollInterval = a.cfg.PollInterval
	var reportInterval = a.cfg.ReportInterval
	creds := insecure.NewCredentials()
	if a.cfg.PublicKeyFile != "" {
		pubKey, err := ciphers.ReadPublicKeyFromFile(a.cfg.PublicKeyFile)
		if err != nil {
			log.Println("loading key error!")
		}
		a.pubKey = pubKey
		log.Println("public key loaded successful.")
	}
	if a.cfg.CertGRPCFile != "" {
		credsTmp, err := credentials.NewClientTLSFromFile(a.cfg.CertGRPCFile, "")
		if err != nil {
			log.Printf("loading GRPC key error: %s", err.Error())
		}
		creds = credsTmp
		log.Println("TLS cert for GRPC loaded.")
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	// 	var rtm runtime.MemStats
	log.Println("Logger start...")
	// Start gRPC client
	fqdn := strings.Split(a.cfg.Server, ":")
	conn, err := grpc.Dial(fqdn[0]+":3200", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// получаем переменную интерфейсного типа UsersClient,
	// через которую будем отправлять сообщения
	c := pb.NewMetricsClient(conn)
	var wg sync.WaitGroup
	tickerCollect := time.NewTicker(pollInterval)
	tickerSend := time.NewTicker(reportInterval)
	defer tickerCollect.Stop()
	defer tickerSend.Stop()
	for {
		select {
		case <-tickerCollect.C:
			go a.CollectLogs()
		case <-tickerSend.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				a.SendLogsbyGRPC(c)
				a.SendLogsbyPost("http://" + a.cfg.Server + "/update/")
				a.SendLogsbyJSON("http://" + a.cfg.Server + "/update/")
				a.SendLogsbyJSONBatch("http://" + a.cfg.Server + "/updates/")
			}()
			wg.Wait()
			//go a.AgentSendWorker(c)
		case <-sigc:
			log.Println("Got quit signal.")
			return
		}
	}
}

// AgentSendWorker - send all collected data by POST,JSON and batch JSON to remote server.
func (a Agent) AgentSendWorker(c pb.MetricsClient) {
	a.SendLogsbyGRPC(c)
	a.SendLogsbyPost("http://" + a.cfg.Server + "/update/")
	a.SendLogsbyJSON("http://" + a.cfg.Server + "/update/")
	a.SendLogsbyJSONBatch("http://" + a.cfg.Server + "/updates/")
	//a.SendLogsbyGRPCBatch(c)
}

func (a Agent) Call(url string, method string, data io.Reader) error {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		log.Printf("New Req error %s", err.Error())
		return err
	}
	ip := a.IP
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Real-IP", ip)
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Req error %s", err.Error())
		return err
	}
	defer response.Body.Close()
	return nil
}

// SendLogsbyJSONBatch - send logs to remote server by JSON batch (fastest way).
func (a Agent) SendLogsbyJSONBatch(url string) error {
	hasher := crypto.NewCryptoService()
	hasher.InitCryptoService(a.cfg.KeyFile)
	var allData = []models.Metrics{}
	for i := range a.logDB {
		data := a.logDB[i]
		if hasher.IsServiceEnable() {
			_, err := hasher.Hash(&data)
			if err != nil {
				log.Println("Hasher error!")
				continue
			}
		}
		allData = append(allData, data)
	}
	jData, _ := json.Marshal(allData)
	if a.pubKey != nil {
		jData = ciphers.EncryptWithPublicKey(jData, a.pubKey)
	}
	err := a.Call(url, "POST", bytes.NewBuffer(jData))
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	log.Println("Sended logs by POST JSON Batch")
	return err
}

// SendLogsbyJSON - send logs to remote server by JSON one by one.
func (a Agent) SendLogsbyJSON(url string) error {
	hasher := crypto.NewCryptoService()
	hasher.InitCryptoService(a.cfg.KeyFile)
	for i := range a.logDB {
		data := a.logDB[i]
		if hasher.IsServiceEnable() {
			_, err := hasher.Hash(&data)
			if err != nil {
				log.Println("Hasher error!")
				continue
			}
		}
		// log.Println(data)
		jData, _ := json.Marshal(data)
		if a.pubKey != nil {
			jData = ciphers.EncryptWithPublicKey(jData, a.pubKey)
		}
		err := a.Call(url, "POST", bytes.NewBuffer(jData))
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
	}
	log.Println("Sended logs by POST JSON")
	// log.Println(a.logDB)
	return nil
}

// SendLogsbyGRPC - send logs to a remote server by gRPC
func (a Agent) SendLogsbyGRPC(c pb.MetricsClient) error {
	md := metadata.New(map[string]string{"X-Real-IP": a.IP})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	hasher := crypto.NewCryptoService()
	hasher.InitCryptoService(a.cfg.KeyFile)
	for i := range a.logDB {
		data := a.logDB[i]
		if hasher.IsServiceEnable() {
			_, err := hasher.Hash(&data)
			if err != nil {
				log.Println("Hasher error!")
				continue
			}
		}
		_, err := c.AddMetric(ctx, &pb.AddMetricRequest{
			Metric: data.ToProto(),
		})
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
	}
	log.Println("Sended logs by GRPC")
	return nil
}

func (a Agent) SendLogsbyGRPCBatch(c pb.MetricsClient) error {
	md := metadata.New(map[string]string{"X-Real-IP": a.IP})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	hasher := crypto.NewCryptoService()
	hasher.InitCryptoService(a.cfg.KeyFile)
	var allData = []*pb.Metric{}
	for i := range a.logDB {
		data := a.logDB[i]
		if hasher.IsServiceEnable() {
			_, err := hasher.Hash(&data)
			if err != nil {
				log.Println("Hasher error!")
				continue
			}

		}
		allData = append(allData, data.ToProto())
	}
	_, err := c.AddMetrics(ctx, &pb.AddMetricsRequest{
		Metrics: allData,
		Size:    int64(len(allData)),
	})
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	log.Println("Sended logs by  GRPC  Batch")
	return err
}

func (a Agent) getPostStrByIndex(i int, url string) string {
	if a.logDB[i].MType == "counter" {
		return fmt.Sprintf(url+"%s/%s/%d", a.logDB[i].MType, a.logDB[i].ID, *a.logDB[i].Delta)
	} else if a.logDB[i].MType == "gauge" {
		return fmt.Sprintf(url+"%s/%s/%f", a.logDB[i].MType, a.logDB[i].ID, *a.logDB[i].Value)
	}
	return "type unknown"
}

// SendLogsbyPost - send logs to remote server one by one as POST request.
func (a *Agent) SendLogsbyPost(sData string) error {
	for i := range a.logDB {
		if err := a.Call(a.getPostStrByIndex(i, sData), "GET", nil); err != nil {
			log.Printf("Error: %s", err.Error())
		}
	}
	log.Println("Sended logs by POST param")
	// log.Println(a.logDB)
	return nil
}

// CollectLogs - collect runtime metrics and save it to storage.
func (a *Agent) CollectLogs() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	virtualMem, err := mem.VirtualMemory()
	utils.CheckError(err)
	CPU, err := cpu.Percent(0, true)
	utils.CheckError(err)
	for i, k := range CPU {
		a.AppendMetric(models.NewMetric(("CPUutilization" + fmt.Sprint(i+1)), "gauge", nil, &k, ""))
	}
	var tmpTM = float64(virtualMem.Total)
	var tmpFM = float64(virtualMem.Free)
	a.AppendMetric(models.NewMetric("TotalMemory", "gauge", nil, &tmpTM, ""))
	a.AppendMetric(models.NewMetric("TotalMemory", "gauge", nil, &tmpFM, ""))
	// log.Println(a.logDB)
	var tmpAlloc = float64(rtm.Alloc)
	a.AppendMetric(models.Metrics{ID: "Alloc", MType: "gauge", Delta: nil, Value: &tmpAlloc})
	var tmpBuckHashSys = float64(rtm.BuckHashSys)
	a.AppendMetric(models.Metrics{ID: "BuckHashSys", MType: "gauge", Delta: nil, Value: &tmpBuckHashSys})
	var tmpFrees = float64(rtm.Frees)
	a.AppendMetric(models.Metrics{ID: "Frees", MType: "gauge", Delta: nil, Value: &tmpFrees})
	var tmpGCCPUFraction = float64(rtm.GCCPUFraction)
	a.AppendMetric(models.Metrics{ID: "GCCPUFraction", MType: "gauge", Delta: nil, Value: &tmpGCCPUFraction})
	var tmpGCSys = float64(rtm.GCSys)
	a.AppendMetric(models.Metrics{ID: "GCSys", MType: "gauge", Delta: nil, Value: &tmpGCSys})
	var tmpHeapAlloc = float64(rtm.HeapAlloc)
	a.AppendMetric(models.Metrics{ID: "HeapAlloc", MType: "gauge", Delta: nil, Value: &tmpHeapAlloc})
	var tmpHeapIdle = float64(rtm.HeapIdle)
	a.AppendMetric(models.Metrics{ID: "HeapIdle", MType: "gauge", Delta: nil, Value: &tmpHeapIdle})
	var tmpHeapInuse = float64(rtm.HeapInuse)
	a.AppendMetric(models.Metrics{ID: "HeapInuse", MType: "gauge", Delta: nil, Value: &tmpHeapInuse})
	var tmpHeapObject = float64(rtm.HeapObjects)
	a.AppendMetric(models.Metrics{ID: "HeapObjects", MType: "gauge", Delta: nil, Value: &tmpHeapObject})
	var tmpHeapReleased = float64(rtm.HeapReleased)
	a.AppendMetric(models.Metrics{ID: "HeapReleased", MType: "gauge", Delta: nil, Value: &tmpHeapReleased})
	var tmpHeapSys = float64(rtm.HeapSys)
	a.AppendMetric(models.Metrics{ID: "HeapSys", MType: "gauge", Delta: nil, Value: &tmpHeapSys})
	var tmpLastGC = float64(rtm.LastGC)
	a.AppendMetric(models.Metrics{ID: "LastGC", MType: "gauge", Delta: nil, Value: &tmpLastGC})
	var tmpLookups = float64(rtm.Lookups)
	a.AppendMetric(models.Metrics{ID: "Lookups", MType: "gauge", Delta: nil, Value: &tmpLookups})
	var tmpMCacheInuse = float64(rtm.MCacheInuse)
	a.AppendMetric(models.Metrics{ID: "MCacheInuse", MType: "gauge", Delta: nil, Value: &tmpMCacheInuse})
	var tmpMCacheSys = float64(rtm.MCacheSys)
	a.AppendMetric(models.Metrics{ID: "MCacheSys", MType: "gauge", Delta: nil, Value: &tmpMCacheSys})
	var tmpMSpanInuse = float64(rtm.MSpanInuse)
	a.AppendMetric(models.Metrics{ID: "MSpanInuse", MType: "gauge", Delta: nil, Value: &tmpMSpanInuse})
	var tmpMSpanSys = float64(rtm.MSpanSys)
	a.AppendMetric(models.Metrics{ID: "MSpanSys", MType: "gauge", Delta: nil, Value: &tmpMSpanSys})
	var tmpNextGC = float64(rtm.NextGC)
	a.AppendMetric(models.Metrics{ID: "NextGC", MType: "gauge", Delta: nil, Value: &tmpNextGC})
	var tmpOtherSys = float64(rtm.OtherSys)
	a.AppendMetric(models.Metrics{ID: "OtherSys", MType: "gauge", Delta: nil, Value: &tmpOtherSys})
	var tmpPauseTotalNs = float64(rtm.PauseTotalNs)
	a.AppendMetric(models.Metrics{ID: "PauseTotalNs", MType: "gauge", Delta: nil, Value: &tmpPauseTotalNs})
	var tmpRandomValue = float64(rand.Int63())
	a.AppendMetric(models.Metrics{ID: "RandomValue", MType: "gauge", Delta: nil, Value: &tmpRandomValue})
	a.counter = a.counter + 1
	tmpI := a.counter
	a.AppendMetric(models.Metrics{ID: "PollCount", MType: "counter", Delta: &tmpI, Value: nil})
	var tmpNumForcedGC = float64(rtm.NumForcedGC)
	a.AppendMetric(models.Metrics{ID: "NumForcedGC", MType: "gauge", Delta: nil, Value: &tmpNumForcedGC})
	var tmpNumGC = float64(rtm.NumGC)
	a.AppendMetric(models.Metrics{ID: "NumGC", MType: "gauge", Delta: nil, Value: &tmpNumGC})
	var tmpStackInuse = float64(rtm.StackInuse)
	a.AppendMetric(models.Metrics{ID: "StackInuse", MType: "gauge", Delta: nil, Value: &tmpStackInuse})
	var tmpStackSys = float64(rtm.StackSys)
	a.AppendMetric(models.Metrics{ID: "StackSys", MType: "gauge", Delta: nil, Value: &tmpStackSys})
	var tmpSys = float64(rtm.Sys)
	a.AppendMetric(models.Metrics{ID: "Sys", MType: "gauge", Delta: nil, Value: &tmpSys})
	var tmpTotalAlloc = float64(rtm.TotalAlloc)
	a.AppendMetric(models.Metrics{ID: "TotalAlloc", MType: "gauge", Delta: nil, Value: &tmpTotalAlloc})
	var tmpMallocs = float64(rtm.Mallocs)
	a.AppendMetric(models.Metrics{ID: "Mallocs", MType: "gauge", Delta: nil, Value: &tmpMallocs})
	var tmpTotalMem = float64(virtualMem.Total)
	a.AppendMetric(models.Metrics{ID: "TotalMemory", MType: "gauge", Delta: nil, Value: &tmpTotalMem})
	var tmpFreeMem = float64(virtualMem.Free)
	a.AppendMetric(models.Metrics{ID: "FreeMemory", MType: "gauge", Delta: nil, Value: &tmpFreeMem})
	a.AppendMetric(models.Metrics{ID: "CPUutilization1", MType: "gauge", Delta: nil, Value: &CPU[0]})
	log.Println("Collected logs")
}

func parseCfg() Config {
	var cfg Config
	var cfgFlag Config
	var envCfg = make(map[string]bool)
	opts := env.Options{
		OnSet: func(tag string, value interface{}, isDefault bool) {
			envCfg[tag] = isDefault
		},
	}
	//  Сначала читаем ключи

	flag.StringVar(&cfgFlag.Server, "a", "localhost:8080", "host:port (default localhost:8080)")
	flag.DurationVar(&cfgFlag.ReportInterval, "r", time.Duration(10*time.Second), "report to server interval in seconds (default 10s)")
	flag.DurationVar(&cfgFlag.PollInterval, "p", time.Duration(2*time.Second), "poll interval in seconds (default 2s)")
	flag.StringVar(&cfgFlag.KeyFile, "k", "", "hmac key")
	flag.StringVar(&cfgFlag.PublicKeyFile, "crypto-key", "", "public key")
	flag.StringVar(&cfg.configFile, "c", "", "json config file path")
	flag.StringVar(&cfg.configFile, "config", "", "json config file path")
	flag.StringVar(&cfgFlag.CertGRPCFile, "cert", "", "path to GRPC auth cert file")
	flag.Parse()
	err := env.Parse(&cfg, opts)
	if err != nil {
		log.Fatal(err)
	}
	if flag := flag.Lookup("a"); (flag != nil) && envCfg["ADDRESS"] {
		cfg.Server = cfgFlag.Server
	}
	if flag := flag.Lookup("r"); (flag != nil) && envCfg["REPORT_INTERVAL"] {
		cfg.ReportInterval = cfgFlag.ReportInterval
	}
	if flag := flag.Lookup("p"); (flag != nil) && envCfg["POLL_INTERVAL"] {
		cfg.PollInterval = cfgFlag.PollInterval
	}
	if flag := flag.Lookup("k"); (flag != nil) && envCfg["KEY"] {
		cfg.KeyFile = cfgFlag.KeyFile
	}
	if flag := flag.Lookup("crypto-key"); (flag != nil) && envCfg["CRYPTO_KEY"] {
		cfg.PublicKeyFile = cfgFlag.PublicKeyFile
	}
	if flag := flag.Lookup("cert"); (flag != nil) && envCfg["CERT_FILE"] {
		cfg.CertGRPCFile = cfgFlag.CertGRPCFile
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
	return cfg
}
