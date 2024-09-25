package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Environment string

const (
	EnvTest       Environment = "test"
	EnvLocal      Environment = "local"
	EnvHomolog    Environment = "homolog"
	EnvSandbox    Environment = "sandbox"
	EnvProduction Environment = "production"
)

type Config struct {
	Environment Environment `required:"true" envconfig:"ENVIRONMENT"`
	Development bool        `required:"true" envconfig:"DEVELOPMENT"`

	App    App
	Server Server

	// Resilience
	CircuitBreaker CircuitBreaker
	Retry          Retry

	// Infra
	Otel     Otel
	Postgres Postgres
	Redis    Redis
}

type App struct {
	Name                    string        `required:"true" envconfig:"APP_NAME"`
	ID                      string        `required:"true" envconfig:"APP_ID"`
	GracefulShutdownTimeout time.Duration `required:"true" envconfig:"APP_GRACEFUL_SHUTDOWN_TIMEOUT"`
}

type Server struct {
	SwaggerHost  string        `required:"true" envconfig:"SERVER_SWAGGER_HOST"`
	APIAddress   string        `required:"true" envconfig:"SERVER_API_ADDRESS"`
	ReadTimeout  time.Duration `required:"true" envconfig:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `required:"true" envconfig:"SERVER_WRITE_TIMEOUT"`
}

type CircuitBreaker struct {
	Timeout time.Duration `required:"true" envconfig:"CIRCUIT_BREAKER_TIMEOUT"`

	SleepWindow time.Duration `required:"true" envconfig:"CIRCUIT_BREAKER_SLEEP_WINDOW"`

	MaxConcurrentRequests int `required:"true" envconfig:"CIRCUIT_BREAKER_MAX_CONCURRENT_REQUESTS"`

	RequestVolumeThreshold int `required:"true" envconfig:"CIRCUIT_BREAKER_REQUEST_VOLUME_THRESHOLD"`

	ErrorPercentThreshold int `required:"true" envconfig:"CIRCUIT_BREAKER_ERROR_PERCENT_THRESHOLD"`
}

type Retry struct {
	MaxAttempts int           `required:"true" envconfig:"RETRY_MAX_ATTEMPTS"`
	WaitMin     time.Duration `required:"true" envconfig:"RETRY_WAIT_MIN"`
	WaitMax     time.Duration `required:"true" envconfig:"RETRY_WAIT_MAX"`
	Timeout     time.Duration `required:"true" envconfig:"RETRY_TIMEOUT"`
}

type Otel struct {
	CollectorEndpoint string        `required:"true" envconfig:"OTEL_COLLECTOR_ENDPOINT"`
	ExporterTimeout   time.Duration `required:"true" envconfig:"OTEL_EXPORTER_TIMEOUT"`

	// The ratio of samples sent by TraceID. See more on TraceIDRatioBased.
	// NOTE: The sampling in production is always 1% (100:1). So just values lesser than 1% will make an effect.
	SamplingRatio    float64 `required:"true" envconfig:"OTEL_SAMPLING_RATIO"` // 0.01 is a 100:1 ratio
	ServiceName      string  `required:"true" envconfig:"OTEL_SERVICE_NAME"`
	ServiceNamespace string  `required:"true" envconfig:"OTEL_SERVICE_NAMESPACE"`
}

type Postgres struct {
	DatabaseName          string `envconfig:"DATABASE_NAME"                    default:"dev"`
	User                  string `envconfig:"DATABASE_USER"                    default:"postgres"`
	Password              string `envconfig:"DATABASE_PASSWORD"                default:"postgres"`
	Host                  string `envconfig:"DATABASE_HOST_DIRECT"             default:"localhost"`
	Port                  string `envconfig:"DATABASE_PORT_DIRECT"             default:"5432"`
	PoolMinSize           string `envconfig:"DATABASE_POOL_MIN_SIZE"           default:"2"`
	PoolMaxSize           string `envconfig:"DATABASE_POOL_MAX_SIZE"           default:"10"`
	PoolMaxConnLifetime   string `envconfig:"DATABASE_POOL_MAX_CONN_LIFETIME"`
	PoolMaxConnIdleTime   string `envconfig:"DATABASE_POOL_MAX_CONN_IDLE_TIME"`
	PoolHealthCheckPeriod string `envconfig:"DATABASE_POOL_HEALTHCHECK_PERIOD"`
	SSLMode               string `envconfig:"DATABASE_SSLMODE"                 default:"disable"`
	SSLRootCert           string `envconfig:"DATABASE_SSL_ROOTCERT"`
	SSLCert               string `envconfig:"DATABASE_SSL_CERT"`
	SSLKey                string `envconfig:"DATABASE_SSL_KEY"`
	Hostname              string `envconfig:"HOSTNAME"`
}

type Redis struct {
	Host     string `required:"true"        envconfig:"REDIS_ADDR"`
	Port     string `required:"true"        envconfig:"REDIS_PORT"`
	User     string `envconfig:"REDIS_USER"`
	Password string `required:"true"        envconfig:"REDIS_PASSWORD"`
	UseTLS   bool   `required:"true"        envconfig:"REDIS_USE_TLS"`
}

// Address returns "host:port" string for connection.
func (r Redis) Address() string {
	return r.Host + ":" + r.Port
}

func New() (Config, error) {
	const operation = "Config.New"

	var cfg Config

	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("%s -> %w", operation, err)
	}

	return cfg, nil
}
