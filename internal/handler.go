package internal

import (
  fmt      "fmt"
  http     "net/http"
  json     "encoding/json"
  log       "github.com/sirupsen/logrus"
  resource "github.com/gpenaud/needys-api-resource/internal/resource"
  mux      "github.com/gorilla/mux"
  sql      "database/sql"
  strconv  "strconv"
)

var handlerLog *log.Entry

func init() {
  handlerLog = log.WithFields(log.Fields{
    "_file": "internal/handler.go",
    "_type": "user",
  })
}

// -------------------------------------------------------------------------- //
// Common functions for handlers

func respondHTTPCodeOnly(w http.ResponseWriter, code int) {
  w.WriteHeader(code)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
  handlerLog.Error(message)
  respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
  response, _ := json.Marshal(payload)
  handlerLog.Debug(fmt.Sprintf("JSON response: %s", response))

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  w.Write(response)
}

// -------------------------------------------------------------------------- //
// Probe handlers

func (a *Application) isHealthy(w http.ResponseWriter, _ *http.Request) {
  payload := map[string]bool{
    "healthy": true,
  }

  if a.Config.LogHealthcheck {
    handlerLog.Debug("sent a GET request on /healthy")
    respondWithJSON(w, http.StatusOK, payload)
  } else {
    respondHTTPCodeOnly(w, http.StatusOK)
  }
}

func (a *Application) isReady(w http.ResponseWriter, _ *http.Request) {
  if a.Config.LogHealthcheck {
    handlerLog.Debug("sent a GET request on /ready")

    if err := a.isDatabaseReachable(); err != nil {
      respondWithError(w, http.StatusInternalServerError, "database is not available")
    } else {
      payload := map[string]interface{}{
        "ready": true,
      }

      respondWithJSON(w, http.StatusOK, payload)
    }
  } else {
    if err := a.isDatabaseReachable(); err != nil {
      respondHTTPCodeOnly(w, http.StatusInternalServerError)
    } else {
      respondHTTPCodeOnly(w, http.StatusOK)
    }
  }
}

// -------------------------------------------------------------------------- //
// Maintenance handlers

const dbInitQuery = `
  CREATE TABLE IF NOT EXISTS resources (
    id SERIAL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    CONSTRAINT resources_pkey PRIMARY KEY (id)
  );

  DELETE FROM resources;
  ALTER SEQUENCE resources_id_seq RESTART WITH 1;

  INSERT INTO resources(type, description) VALUES('individual', 'faire une sieste') RETURNING id;
  INSERT INTO resources(type, description) VALUES('collective', 'faire une séance de biodanza') RETURNING id;
  `

func (a *Application) InitializeDB(w http.ResponseWriter, _ *http.Request) {
  var err error

  if err = a.isDatabaseReachable(); err != nil {
    respondWithError(w, http.StatusInternalServerError, "Database is not available")
  } else {
    if _, err = a.DB.Exec(dbInitQuery); err == nil {
      payload := map[string]bool{
        "initialized": true,
      }
      respondWithJSON(w, http.StatusOK, payload)
    } else {
      respondWithError(w, http.StatusInternalServerError, "Database is not initializable")
    }
  }
}

// -------------------------------------------------------------------------- //
// Resource handlers

func (a *Application) getResources(w http.ResponseWriter, r *http.Request) {
  handlerLog.Info("sent a GET query on /resources")

  count, _ := strconv.Atoi(r.FormValue("count"))
  start, _ := strconv.Atoi(r.FormValue("start"))

  if count > 10 || count < 1 {
    count = 10
  }

  if start < 0 {
    start = 0
  }

  products, err := resource.GetResources(a.DB, start, count)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, products)
}

func (a *Application) getResource(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  handlerLog.WithFields(log.Fields{
    "parameter_id": vars["id"],
  }).Info("sent a GET query on /resource/{id}")

  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, fmt.Sprintf("The resource with ID %d is invalid", id))
    return
  }

  resource := resource.Resource{ID: id}

  err = resource.GetResource(a.DB)
  if err != nil {
    switch err {
    case sql.ErrNoRows:
      respondWithError(w, http.StatusNotFound, fmt.Sprintf("The resource with ID %d is not found", id))
    default:
      respondWithError(w, http.StatusInternalServerError, err.Error())
    }
    return
  }

  respondWithJSON(w, http.StatusOK, resource)
}

func (a *Application) createResource(w http.ResponseWriter, r *http.Request) {
  handlerLog.Info("sent a POST query on /resource to create a new resource")

  var resource resource.Resource

  decoder := json.NewDecoder(r.Body)

  err := decoder.Decode(&resource)
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "The payload is invalid")
    return
  }

  defer r.Body.Close()

  err = resource.CreateResource(a.DB)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusCreated, resource)
}

func (a *Application) updateResource(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  handlerLog.WithFields(log.Fields{
    "parameter_id": vars["id"],
  }).Info("sent a PUT query on /resource/{id} to update the resource")

  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "The resource ID is invalid")
    return
  }

  var resource resource.Resource
  decoder := json.NewDecoder(r.Body)

  err = decoder.Decode(&resource)
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "The payload is invalid")
    return
  }

  defer r.Body.Close()

  resource.ID = id

  err = resource.UpdateResource(a.DB)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, resource)
}

func (a *Application) deleteResource(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  handlerLog.WithFields(log.Fields{
    "parameter_id": vars["id"],
  }).Info("sent a DELETE query on /resource/{id} to delete the resource")

  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "The resource ID is invalid")
    return
  }

  resource := resource.Resource{ID: id}

  err = resource.DeleteResource(a.DB)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
