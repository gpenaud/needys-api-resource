package internal

import (
  context "context"
  fmt  "fmt"
  http "net/http"
  log  "github.com/sirupsen/logrus"
  _    "github.com/lib/pq"
  mux  "github.com/gorilla/mux"
  sql  "database/sql"
  time "time"
)

var applicationLog *log.Entry

func init() {
  applicationLog = log.WithFields(log.Fields{
    "_file": "internal/application.go",
    "_type": "system",
  })
}

type Configuration struct {
  Environment    string
  Verbosity      string
  LogFormat      string
  LogHealthcheck bool
  Server struct {
    Host string
    Port string
  }
  Database struct {
    Host     string
    Port     string
    Name     string
    Username string
    Password string
  }
  Healthcheck struct {
    Timeout  int
  }
}

type Version struct {
  BuildTime string
  Commit    string
  Release   string
}

type Application struct {
  Router  *mux.Router
  DB      *sql.DB
  Config  *Configuration
  Version *Version
}

func (a *Application) Initialize() {
  connectionString :=
    fmt.Sprintf(
      "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
      a.Config.Database.Host,
      a.Config.Database.Port,
      a.Config.Database.Username,
      a.Config.Database.Password,
      a.Config.Database.Name,
    )

  var err error

  if a.DB, err = sql.Open("postgres", connectionString); err != nil {
    applicationLog.Error(err)
  }

  a.Router = mux.NewRouter()

  a.initializeLogger()
  a.initializeRoutes()

  applicationLog.WithFields(log.Fields{
    "database_host": a.Config.Database.Host,
    "database_port": a.Config.Database.Port,
    "database_username": a.Config.Database.Username,
    "database_name":   a.Config.Database.Name,
  }).Info("trying to connect to database")

  applicationLog.Info("application is initialized")
}

func (a *Application) initializeLogger() {
  switch a.Config.Verbosity {
  case "fatal":
    log.SetLevel(log.FatalLevel)
  case "error":
    log.SetLevel(log.ErrorLevel)
  case "warning":
    log.SetLevel(log.WarnLevel)
  case "info":
    log.SetLevel(log.InfoLevel)
  case "debug":
    log.SetLevel(log.DebugLevel)
    log.SetReportCaller(false)
  default:
    log.WithFields(
      log.Fields{"verbosity": a.Config.Verbosity},
    ).Fatal("Unkown verbosity level")
  }

  switch a.Config.Environment {
  case "development":
    log.SetFormatter(&log.TextFormatter{})
  case "integration":
    log.SetFormatter(&log.JSONFormatter{})
  case "production":
    log.SetFormatter(&log.JSONFormatter{})
  default:
    log.WithFields(
      log.Fields{"environment": a.Config.Environment},
    ).Fatal("Unkown environment type")
  }

  if a.Config.LogFormat != "unset" {
    switch a.Config.LogFormat {
    case "text":
      log.SetFormatter(&log.TextFormatter{})
    case "json":
      log.SetFormatter(&log.JSONFormatter{})
    default:
      log.WithFields(
        log.Fields{"log_format": a.Config.LogFormat},
      ).Fatal("Unkown log format")
    }
  }
}

func (a *Application) initializeRoutes() {
  // application resource-related routes
  a.Router.HandleFunc("/resources", a.getResources).Methods("GET")
  a.Router.HandleFunc("/resource", a.createResource).Methods("POST")
  a.Router.HandleFunc("/resource/{id:[0-9]+}", a.getResource).Methods("GET")
  a.Router.HandleFunc("/resource/{id:[0-9]+}", a.updateResource).Methods("PUT")
  a.Router.HandleFunc("/resource/{id:[0-9]+}", a.deleteResource).Methods("DELETE")
  // application probes routes
  a.Router.HandleFunc("/health", a.isHealthy).Methods("GET")
  a.Router.HandleFunc("/ready", a.isReady).Methods("GET")
  // application maintenance routes
  a.Router.HandleFunc("/initialize_db", a.InitializeDB).Methods("GET")
}

func (a *Application) Run(ctx context.Context) {
  server_address :=
    fmt.Sprintf("%s:%s", a.Config.Server.Host, a.Config.Server.Port)

  server_message :=
    fmt.Sprintf(
  `

START INFORMATIONS
------------------
Listening needys-api-resource on %s:%s...

BUILD INFORMATIONS
------------------
time: %s
release: %s
commit: %s

`,
      a.Config.Server.Host,
      a.Config.Server.Port,
      a.Version.BuildTime,
      a.Version.Release,
      a.Version.Commit,
    )

  httpServer := &http.Server{
		Addr:    server_address,
		Handler: a.Router,
	}

  go func() {
    // we keep this log on standard format
    log.Info(server_message)
    applicationLog.Fatal(httpServer.ListenAndServe())
  }()

  <-ctx.Done()

  applicationLog.Info("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

  var err error

	if err = httpServer.Shutdown(ctxShutDown); err != nil {
    applicationLog.WithFields(log.Fields{
      "error": err,
    }).Fatal("server shutdown failed")
	}

  applicationLog.Info("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}
