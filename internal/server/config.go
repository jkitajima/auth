package server

import (
	"fmt"
	"io"
	"strings"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffyaml"
)

type Environment string

const (
	EnvironmentTest  Environment = "test"
	EnvironmentLocal Environment = "local"
)

func NewEnvironment(env string) Environment {
	switch env {
	case "test":
		return EnvironmentTest
	case "local":
		fallthrough
	default:
		return EnvironmentLocal
	}
}

type Config struct {
	Environment Environment
	Server      *Server
	Auth        *Auth
	DB          *DB
}

type Server struct {
	Host           string
	Port           string
	Timeout        *Timeout
	Health         *Health
	MaxHeaderBytes int
}

type Timeout struct {
	Read     int
	Write    int
	Idle     int
	Shutdown int
}

type Health struct {
	Timeout  int
	Cache    int
	Interval int
	Delay    int
	Retries  int
}

type Auth struct {
	JWT *JWT
}

type JWT struct {
	Algorithm  string
	Key        string
	Issuer     string
	Audience   []string
	Expiration int
}

type DB struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSL      string
	DSN      string
}

func NewConfig(stdout io.Writer, args []string) (*Config, error) {
	fs := ff.NewFlagSet(Service)
	var (
		config                string
		env                   string
		serverHost            string
		serverPort            string
		serverTimeoutRead     int
		serverTimeoutWrite    int
		serverTimeoutIdle     int
		serverTimeoutShutdown int
		serverHealthTimeout   int
		serverHealthCache     int
		serverHealthInterval  int
		serverHealthDelay     int
		serverHealthRetries   int
		serverMaxHeaderBytes  int
		authJWTAlg            string
		authJWTKey            string
		authJWTIssuer         string
		authJWTAudience       []string
		authJWTExpiration     int
		dbHost                string
		dbPort                string
		dbName                string
		dbUser                string
		dbPasswd              string
		dbSSL                 string
	)
	fs.StringEnumVar(&config, 0, "config", "environment configuration file", "env.local.yaml", "env.test.yaml")
	fs.StringEnumVar(&env, 0, "env", "build environment", string(EnvironmentLocal), string(EnvironmentTest))
	fs.StringVar(&serverHost, 0, "server.host", "localhost", "server host address to listen for incoming requests")
	fs.StringVar(&serverPort, 0, "server.port", "8080", "server port number to listen for incoming requests")
	fs.IntVar(&serverTimeoutRead, 0, "server.timeout.read", 5, "number of seconds that the server will wait for reading requests")
	fs.IntVar(&serverTimeoutWrite, 0, "server.timeout.write", 5, "number of seconds that the server will wait for writing requests")
	fs.IntVar(&serverTimeoutIdle, 0, "server.timeout.idle", 120, "number of seconds that the server will wait for the next request")
	fs.IntVar(&serverTimeoutShutdown, 0, "server.timeout.shutdown", 30, "the duration for which the server gracefully wait for existing connections to finish")
	fs.IntVar(&serverHealthTimeout, 0, "server.health.timeout", 30, "if a single run of the check takes longer than timeout seconds then the check is considered to have failed")
	fs.IntVar(&serverHealthCache, 0, "server.health.cache", 5, "sets the duration for how long the aggregated health check result will be cached")
	fs.IntVar(&serverHealthInterval, 0, "server.health.interval", 30, "the health check will first run interval seconds after the program is started, and then again interval seconds after each previous check completes")
	fs.IntVar(&serverHealthDelay, 0, "server.health.delay", 5, "the initialization time for the program to bootstrap before the health check begins")
	fs.IntVar(&serverHealthRetries, 0, "server.health.retries", 3, "the number of consecutive failures of the health check for the container to be considered unhealthy")
	fs.IntVar(&serverMaxHeaderBytes, 0, "server.header", 10240, "number of bytes that will be the maximum permitted size of the headers in an HTTP request")
	fs.StringVar(&authJWTAlg, 0, "auth.jwt.alg", "HS256", "algorithm that was used for signing the JWT token")
	fs.StringVar(&authJWTKey, 0, "auth.jwt.key", "", "key that was used for signing the JWT token")
	fs.StringVar(&authJWTIssuer, 0, "auth.jwt.iss", "", `the "iss" (issuer) claim identifies the principal that issued the jwt`)
	fs.StringListVar(&authJWTAudience, 0, "auth.jwt.aud", `the "aud" (audience) claim identifies the recipients that the jwt is intended for`)
	fs.IntVar(&authJWTExpiration, 0, "auth.jwt.exp", 1200, `the "exp" (expiration time) claim identifies the expiration time on or after which the jwt must not be accepted for processing`)
	fs.StringVar(&dbHost, 0, "db.host", "", "database host address")
	fs.StringVar(&dbPort, 0, "db.port", "", "database port number")
	fs.StringVar(&dbName, 0, "db.name", "", "database name")
	fs.StringVar(&dbUser, 0, "db.user", "", "database user")
	fs.StringVar(&dbPasswd, 0, "db.passwd", "", "database password")
	fs.StringVar(&dbSSL, 0, "db.ssl", "", "database ssl mode")

	if err := ff.Parse(fs, args[1:],
		ff.WithEnvVarPrefix(strings.ToUpper(Service)),
		ff.WithConfigFileParser(ffyaml.Parse),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigIgnoreUndefinedFlags(),
	); err != nil {
		fmt.Fprintf(stdout, "%s\n", ffhelp.Flags(fs))
		fmt.Fprintf(stdout, "ERROR\n%v\n", err)
		return &Config{}, err
	}

	return &Config{
		Environment: NewEnvironment(env),
		Server: &Server{
			Host: serverHost,
			Port: serverPort,
			Timeout: &Timeout{
				Read:     serverTimeoutRead,
				Write:    serverTimeoutWrite,
				Idle:     serverTimeoutIdle,
				Shutdown: serverTimeoutShutdown,
			},
			Health: &Health{
				Timeout:  serverHealthTimeout,
				Cache:    serverHealthCache,
				Interval: serverHealthInterval,
				Delay:    serverHealthDelay,
				Retries:  serverHealthRetries,
			},
			MaxHeaderBytes: serverMaxHeaderBytes,
		},
		Auth: &Auth{
			&JWT{
				Algorithm:  authJWTAlg,
				Key:        authJWTKey,
				Issuer:     authJWTIssuer,
				Audience:   authJWTAudience,
				Expiration: authJWTExpiration,
			},
		},
		DB: &DB{
			Host:     dbHost,
			Port:     dbPort,
			Name:     dbName,
			User:     dbUser,
			Password: dbPasswd,
			SSL:      dbSSL,
		},
	}, nil
}
