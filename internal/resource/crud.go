package resource

import (
  log  "github.com/sirupsen/logrus"
  sql "database/sql"
)

type Resource struct {
  ID          int    `json:"id"`
  Type        string `json:"type"`
  Description string `json:"description"`
}

func (r *Resource) GetResource(db *sql.DB) error {
  log.WithFields(log.Fields{
    "scope": "user",
    "type": "database query",
    "parameter_id": r.ID,
  }).Debug("SELECT type, description FROM resources WHERE id=%d")

  return db.QueryRow("SELECT type, description FROM resources WHERE id=$1",
    r.ID).Scan(&r.Type, &r.Description)
}

func (r *Resource) UpdateResource(db *sql.DB) error {
  log.WithFields(log.Fields{
    "scope": "user",
    "type": "database query",
    "parameter_type": r.Type,
    "parameter_description": r.Description,
    "parameter_id": r.ID,
  }).Debug("UPDATE resources SET type=%s, description=%s WHERE id=%d")

  _, err :=
    db.Exec("UPDATE resources SET type=$1, description=$2 WHERE id=$3",
      r.Type, r.Description, r.ID)

  return err
}

func (r *Resource) DeleteResource(db *sql.DB) error {
  log.WithFields(log.Fields{
    "scope": "user",
    "type": "database query",
    "parameter_id": r.ID,
  }).Debug("DELETE FROM resources WHERE id=%d")

  _, err := db.Exec("DELETE FROM resources WHERE id=$1", r.ID)

  return err
}

func (r *Resource) CreateResource(db *sql.DB) error {
  log.WithFields(log.Fields{
    "scope": "user",
    "type": "database query",
    "parameter_type": r.Type,
    "parameter_description": r.Description,
    "parameter_id": r.ID,
  }).Debug("INSERT INTO resources(type, description) VALUES(%s, %s) RETURNING id")

  err := db.QueryRow(
    "INSERT INTO resources(type, description) VALUES($1, $2) RETURNING id",
    r.Type, r.Description).Scan(&r.ID)

  if err != nil {
    return err
  }

  return nil
}

func GetResources(db *sql.DB, start, count int) ([]Resource, error) {
  log.WithFields(log.Fields{
    "scope": "user",
    "type": "database query",
    "parameter_count": count,
    "parameter_start": start,
  }).Debug("SELECT id, type, description FROM resources LIMIT %d OFFSET %d")

  rows, err := db.Query(
    "SELECT id, type, description FROM resources LIMIT $1 OFFSET $2",
    count, start)

  if err != nil {
    return nil, err
  }

  defer rows.Close()

  resources := []Resource{}

  for rows.Next() {
    var r Resource
    if err := rows.Scan(&r.ID, &r.Type, &r.Description); err != nil {
      return nil, err
    }
    resources = append(resources, r)
  }

  return resources, nil
}
