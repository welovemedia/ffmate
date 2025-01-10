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
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/welovemedia/ffmate/docs"
	"github.com/welovemedia/ffmate/sev/metrics"
	"github.com/welovemedia/ffmate/sev/validate"
	"github.com/yosev/debugo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Sev struct {
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

var debug = debugo.New("sev")

func New(name string, version string, dbPath string, port uint) *Sev {
	// setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05.000",
	})
	logger.SetLevel(logrus.InfoLevel)

	// setup gin
	gin.SetMode(gin.ReleaseMode)
	ginInstance := gin.New()

	// setup swagger
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host += fmt.Sprintf(":%d", port)
	ginInstance.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// seutp db
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	if err != nil {
		logger.Errorf("failed to initialize database connection (path: %s): %v", dbPath, err)
		os.Exit(1)
	} else {
		debug.Debugf("initialized database connection (path: %s)", dbPath)
	}

	metrics := &metrics.Metrics{Logger: logger}
	metrics.Init()

	sev := &Sev{
		appStartTime: time.Now(),

		logger: logger,

		metrics: metrics,

		db: db,

		gin:      ginInstance,
		validate: &validate.Validate{},

		ctx: context.Background(),
	}

	sev.registerMetrics()

	return sev
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
		debug.Debug("received interrupt signal, running shutdown hooks")
		for _, hook := range s.shutdownHooks {
			hook(s)
		}
		debug.Debug("shutting down")
		os.Exit(0)
	}()
}

var debugMiddleware = debug.Extend("middleware")

func (s *Sev) RegisterMiddleware(name string, fn func(c *gin.Context, s *Sev)) {
	s.gin.Use(func(c *gin.Context) {
		fn(c, s)
	})

	debugMiddleware.Debugf("registered middleware '%s'", name)
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
