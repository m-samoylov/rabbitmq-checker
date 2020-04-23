package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/labstack/gommon/log"
	"github.com/namsral/flag"
	"github.com/valyala/fasthttp"
)

type NodeStatus struct {
	NodeAvailable    bool
	Timestamp        int64
	HTTPResponseText string
	HTTPResponseCode int
}

type Config struct {
	WebListen         string
	WebReadTimeout    int
	WebWriteTimeout   int
	CheckForceEnabled bool
	CheckInterval     int64
	CheckFailTimeout  int64
	RabbitMQHost      string
	RabbitMQPort      int
	RabbitMQBasicAuth string
	Debug             bool
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	status  = &NodeStatus{}
	config  *Config
)

const (
	keepAliveTimeout int = 60
)

func main() {
	config = parseFlags()

	go checker(status)

	router := getRouter()
	server := &fasthttp.Server{
		Handler:          router.Handler,
		DisableKeepalive: true,
		Concurrency:      2048,
		ReadTimeout:      time.Duration(config.WebReadTimeout) * time.Millisecond,
		WriteTimeout:     time.Duration(config.WebWriteTimeout) * time.Millisecond,
	}

	ln, err := net.Listen("tcp", config.WebListen)
	if err != nil {
		log.Fatalf("Error in net.Listen: %s", err)
	}

	log.Printf("Server starting on %s", config.WebListen)
	if err := server.Serve(ln); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func getRouter() *fasthttprouter.Router {
	router := fasthttprouter.New()
	router.GET("/", checkerHandler)
	router.HEAD("/", checkerHandler)
	return router
}

func parseFlags() *Config {
	var versionFlag bool
	config := Config{}

	flag.StringVar(&config.WebListen, "WEB_LISTEN", ":9672", "Web server listening interface and port")
	flag.IntVar(&config.WebReadTimeout, "WEB_READ_TIMEOUT", 30000, "Web server request read timeout, ms")
	flag.IntVar(&config.WebWriteTimeout, "WEB_WRITE_TIMEOUT", 30000, "Web server request write timeout, ms")
	flag.BoolVar(&config.CheckForceEnabled, "CHECK_FORCE_ENABLED", false, "Ignoring the status of the checks and always marking the node as available")
	flag.Int64Var(&config.CheckInterval, "CHECK_INTERVAL", 1000, "RabbitMQ checks interval, ms")
	flag.Int64Var(&config.CheckFailTimeout, "CHECK_FAIL_TIMEOUT", 3000, "Mark the node inaccessible if for the specified time there were no successful")
	flag.StringVar(&config.RabbitMQHost, "RABBITMQ_HOST", "127.0.0.1", "RabbitMQ host addr")
	flag.IntVar(&config.RabbitMQPort, "RABBITMQ_WEB_PORT", 15672, "RabbitMQ management port")
	flag.StringVar(&config.RabbitMQBasicAuth, "RABBITMQ_BASIC_AUTH", "", "RabbitMQ management Basic Auth")
	flag.BoolVar(&config.Debug, "DEBUG", false, "Debug logs")

	flag.BoolVar(&versionFlag, "version", false, "Show program version")
	if versionFlag {
		fmt.Printf("Version: %s\nGit commit: %s\nBuilding date: %s\n", version, commit, date)
		os.Exit(0)
	}

	flag.Parse()
	return &config
}
