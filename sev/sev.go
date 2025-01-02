package sev

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/welovemedia/ffmate/sev/metrics"
	"github.com/welovemedia/ffmate/sev/validate"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Sev struct {
	appName      string
	appVersion   string
	appStartTime time.Time

	logger *logrus.Logger

	metrics *metrics.Metrics

	db *gorm.DB

	gin           *gin.Engine
	validate      *validate.Validate
	startupHooks  []func(*Sev)
	shutdownHooks []func(*Sev)

	sigChannel chan os.Signal

	ctx context.Context
}

func New(name string, version string, dbPath string, debug bool) *Sev {
	// setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	// setup gin
	gin.SetMode(gin.ReleaseMode)

	// seutp db
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	if err != nil {
		logger.Errorf("failed to initialize database connection (path: %s): %v", dbPath, err)
		os.Exit(1)
	} else {
		logger.Debugf("initialized database connection (path: %s)", dbPath)
	}

	metrics := &metrics.Metrics{Logger: logger}
	metrics.Init()

	sev := &Sev{
		appName:      name,
		appVersion:   version,
		appStartTime: time.Now(),

		logger: logger,

		metrics: metrics,

		db: db,

		gin:      gin.New(),
		validate: &validate.Validate{},

		ctx: context.Background(),
	}

	sev.registerMetrics()

	return sev
}

func (s *Sev) AppName() string {
	return s.appName
}

func (s *Sev) AppVersion() string {
	return s.appVersion
}

func (s *Sev) AppStartTime() time.Time {
	return s.appStartTime
}

func (s *Sev) Logger() *logrus.Logger {
	return s.logger
}

func (s *Sev) Gin() *gin.Engine {
	return s.gin
}

func (s *Sev) DB() *gorm.DB {
	return s.db
}

func (s *Sev) Validate() *validate.Validate {
	return s.validate
}

func (s *Sev) RegisterSignalHook() {
	s.sigChannel = make(chan os.Signal, 64)
	signal.Notify(s.sigChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-s.sigChannel
		s.logger.Debug("received interrupt signal, running shutdown hooks")
		for _, hook := range s.shutdownHooks {
			hook(s)
		}
		s.logger.Debug("shutting down")
		os.Exit(0)
	}()
}

func (s *Sev) RegisterMiddleware(name string, fn func(c *gin.Context, s *Sev)) {
	s.gin.Use(func(c *gin.Context) {
		fn(c, s)
	})
	s.logger.Debugf("registered middleware '%s'", name)
}

func (s *Sev) RegisterStartupHook(hook func(*Sev)) {
	s.startupHooks = append(s.startupHooks, hook)
}
func (s *Sev) RegisterShutdownHook(hook func(*Sev)) {
	s.shutdownHooks = append(s.shutdownHooks, hook)
}

func (s *Sev) Start(port uint) error {
	for _, hook := range s.startupHooks {
		hook(s)
	}

	return s.gin.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
