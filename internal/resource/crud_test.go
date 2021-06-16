package resource_test

import (
  bytes    "bytes"
  http     "net/http"
  httptest "net/http/httptest"
  internal "github.com/gpenaud/needys-api-resource/internal"
  json     "encoding/json"
  log      "log"
  os       "os"
  testing  "testing"
  // strconv  "strconv"
  // bytes    "bytes"
 )

const tableCreationQuery = `
  CREATE TABLE IF NOT EXISTS resources (
    id SERIAL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    CONSTRAINT resources_pkey PRIMARY KEY (id)
  )`

var a internal.Application

func TestMain(m *testing.M) {

  a.Config = &internal.Configuration{}
  a.Version = &internal.Version{}

  a.Config.Database.Host     = "0.0.0.0"
  a.Config.Database.Port     = "5432"
  a.Config.Database.Name     = "postgres"
  a.Config.Database.Username = "postgres"
  a.Config.Database.Password = "postgres"

  a.Initialize()
  ensureTableExists()

  code := m.Run()

  clearTable()
  os.Exit(code)
}

func ensureTableExists() {
  if _, err := a.DB.Exec(tableCreationQuery); err != nil {
    log.Fatal(err)
  }
}

func clearTable() {
  a.DB.Exec("DELETE FROM resources")
  a.DB.Exec("ALTER SEQUENCE resources_id_seq RESTART WITH 1")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
  rr := httptest.NewRecorder()
  a.Router.ServeHTTP(rr, req)

  return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
  if expected != actual {
    t.Errorf(a.I18n().UnxpectedResponseCode, expected, actual)
  }
}

func addResource() {
  a.DB.Exec("INSERT INTO resources(type, description) VALUES('place', 'aller me balader sur le sentier')")
}

func TestEmptyTable(t *testing.T) {
  clearTable()

  req, _   := http.NewRequest("GET", "/resources", nil)
  response := executeRequest(req)

  checkResponseCode(t, http.StatusOK, response.Code)

  if body := response.Body.String(); body != "[]" {
    t.Errorf(a.I18n().UnexpectedNonEmptyArray, body)
  }
}

func TestGetNonExistentResource(t *testing.T) {
  clearTable()

  req, _   := http.NewRequest("GET", "/resource/11", nil)
  response := executeRequest(req)

  checkResponseCode(t, http.StatusNotFound, response.Code)

  var m map[string]string
  json.Unmarshal(response.Body.Bytes(), &m)

  if m["error"] != "Resource not found" {
    t.Errorf("Expected the 'error' key of the response to be set to 'Resource not found'. Got '%s'", m["error"])
  }
}

func TestCreateResource(t *testing.T) {
  clearTable()

  var jsonStr = []byte(`{"type":"human", "description": "appeler yann pour discuter"}`)
  req, _ := http.NewRequest("POST", "/resource", bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  response := executeRequest(req)
  checkResponseCode(t, http.StatusCreated, response.Code)

  var m map[string]interface{}
  json.Unmarshal(response.Body.Bytes(), &m)

  if m["type"] != "human" {
    t.Errorf("Expected resource type to be 'human'. Got '%v'", m["type"])
  }

  if m["description"] != "appeler yann pour discuter" {
    t.Errorf("Expected resource description to be 'appeler yann pour discuter'. Got '%v'", m["description"])
  }

  // the id is compared to 1.0 because JSON unmarshaling converts numbers to
  // floats, when the target is a map[string]interface{}
  if m["id"] != 1.0 {
    t.Errorf("Expected resource ID to be '1'. Got '%v'", m["id"])
  }
}

func TestGetResource(t *testing.T) {
  clearTable()
  addResource()

  req, _ := http.NewRequest("GET", "/resource/1", nil)
  response := executeRequest(req)

  checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateResource(t *testing.T) {
  clearTable()
  addResource()

  req, _   := http.NewRequest("GET", "/resource/1", nil)
  response := executeRequest(req)

  var originalResource map[string]interface{}
  json.Unmarshal(response.Body.Bytes(), &originalResource)

  var jsonStr = []byte(`{"type":"place", "description": "aller me balader sur le chemin"}`)
  req, _ = http.NewRequest("PUT", "/resource/1", bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  response = executeRequest(req)
  checkResponseCode(t, http.StatusOK, response.Code)

  var m map[string]interface{}
  json.Unmarshal(response.Body.Bytes(), &m)

  if m["id"] != originalResource["id"] {
      t.Errorf(a.I18n().UnexpectedIDNotRemainTheSame, originalResource["id"], m["id"])
  }

  if m["type"] != originalResource["type"] {
      t.Errorf(a.I18n().UnexpectedTypeNotChanged, originalResource["type"], m["type"], m["type"])
  }

  if m["description"] == originalResource["description"] {
      t.Errorf(a.I18n().UnexpectedDescriptionNotChanged, originalResource["description"], m["description"], m["description"])
  }
}

func TestDeleteResource(t *testing.T) {
  clearTable()
  addResource()

  req, _ := http.NewRequest("GET", "/resource/1", nil)
  response := executeRequest(req)
  checkResponseCode(t, http.StatusOK, response.Code)

  req, _ = http.NewRequest("DELETE", "/resource/1", nil)
  response = executeRequest(req)
  checkResponseCode(t, http.StatusOK, response.Code)

  req, _ = http.NewRequest("GET", "/resource/1", nil)
  response = executeRequest(req)
  checkResponseCode(t, http.StatusNotFound, response.Code)
}
