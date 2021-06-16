package internal

import (
  http     "net/http"
  json     "encoding/json"
  resource "github.com/gpenaud/needys-api-resource/internal/resource"
  mux      "github.com/gorilla/mux"
  sql      "database/sql"
  strconv  "strconv"
)

// -------------------------------------------------------------------------- //
// Common functions for handlers

func respondWithError(w http.ResponseWriter, code int, message string) {
  respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
  response, _ := json.Marshal(payload)

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  w.Write(response)
}

// -------------------------------------------------------------------------- //
// Probe handlers

func (a *Application) isHealthy(w http.ResponseWriter, _ *http.Request) {
  respondWithJSON(w, http.StatusOK, "{'healthy': 'true'}")
}

func (a *Application) isReady(w http.ResponseWriter, _ *http.Request) {
  if err := a.isDatabaseReachable(); err != nil {
    respondWithError(w, http.StatusInternalServerError, a.I18n().DatabaseNotAvailable)
  } else {
    respondWithJSON(w, http.StatusOK, "{'ready': 'true'}")
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
    respondWithError(w, http.StatusInternalServerError, a.I18n().DatabaseNotAvailable)
  } else {
    if _, err = a.DB.Exec(dbInitQuery); err == nil {
      respondWithJSON(w, http.StatusOK, "{'initialize': 'true'}")
    } else {
      respondWithError(w, http.StatusInternalServerError, a.I18n().DatabaseNotInitializable)
    }
  }
}

// -------------------------------------------------------------------------- //
// Resource handlers

func (a *Application) getResources(w http.ResponseWriter, r *http.Request) {
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

  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, a.I18n().InvalidResourceID)
    return
  }

  resource := resource.Resource{ID: id}

  err = resource.GetResource(a.DB)
  if err != nil {
    switch err {
    case sql.ErrNoRows:
      respondWithError(w, http.StatusNotFound, a.I18n().ResourceNotfound)
    default:
      respondWithError(w, http.StatusInternalServerError, err.Error())
    }
    return
  }

  respondWithJSON(w, http.StatusOK, resource)
}

func (a *Application) createResource(w http.ResponseWriter, r *http.Request) {
  var resource resource.Resource

  decoder := json.NewDecoder(r.Body)

  err := decoder.Decode(&resource)
  if err != nil {
    respondWithError(w, http.StatusBadRequest, a.I18n().InvalidPayloadRequest)
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

  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, a.I18n().InvalidResourceID)
    return
  }

  var resource resource.Resource
  decoder := json.NewDecoder(r.Body)

  err = decoder.Decode(&resource)
  if err != nil {
    respondWithError(w, http.StatusBadRequest, a.I18n().InvalidPayloadRequest)
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

  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, a.I18n().InvalidResourceID)
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
