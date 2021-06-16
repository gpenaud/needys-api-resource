package internal

import (
  fmt  "fmt"
  http "net/http"
  log  "log"
  _    "github.com/lib/pq"
  mux  "github.com/gorilla/mux"
  sql  "database/sql"
)

type Configuration struct {
  Language string
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
    log.Println(err)
  }

  a.Router = mux.NewRouter()
  a.initializeRoutes()
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
      a.I18n().StartingApplicationMessage,
      a.Config.Server.Host,
      a.Config.Server.Port,
      a.Version.BuildTime,
      a.Version.Release,
      a.Version.Commit,
    )

  log.Println(server_message)
  log.Fatal(http.ListenAndServe(server_address, a.Router))
}
