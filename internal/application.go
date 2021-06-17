package internal

import (
  fmt  "fmt"
  http "net/http"
  log  "github.com/sirupsen/logrus"
  _    "github.com/lib/pq"
  mux  "github.com/gorilla/mux"
  sql  "database/sql"
)

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
  log.Info("system - application is initializing...")
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
    log.Error(err)
  }

  a.Router = mux.NewRouter()

  a.initializeLogger()
  a.initializeRoutes()

  log.WithFields(log.Fields{
    "database_host": a.Config.Database.Host,
    "database_port": a.Config.Database.Port,
    "database_username": a.Config.Database.Username,
    "database_name":   a.Config.Database.Name,
  }).Info("system - trying to connect to database")

  log.Print("system - application initialized !")
}

func (a *Application) initializeLogger() {
  switch a.Config.Verbosity {
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

func (a *Application) Run() {
  server_address :=
    fmt.Sprintf("%s:%s", a.Config.Server.Host, a.Config.Server.Port)

  server_message :=
    fmt.Sprintf(
      "Starting needys-api-resource on %s:%s...\n > build time: %s\n > release: %s\n > commit: %s\n",
      a.Config.Server.Host,
      a.Config.Server.Port,
      a.Version.BuildTime,
      a.Version.Release,
      a.Version.Commit,
    )

  log.Info(server_message)
  log.Fatal(http.ListenAndServe(server_address, a.Router))
}
