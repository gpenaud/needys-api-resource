package main

import (
  cmdline  "github.com/galdor/go-cmdline"
  internal "github.com/gpenaud/needys-api-resource/internal"
  os       "os"
)

func registerCliConfiguration(a *internal.Application) {
  cmdline := cmdline.New()

  a.Config = &internal.Configuration{}

  // application configuration flags
  cmdline.AddOption("", "language", "LANGUAGE", "language of application (EN/FR ...)")
  cmdline.SetOptionDefault("language", "EN")
  // application server configuration flags
  cmdline.AddOption("", "server.host", "HOST", "host of application")
  cmdline.SetOptionDefault("server.host", "localhost")
  cmdline.AddOption("", "server.port", "PORT", "port of application")
  cmdline.SetOptionDefault("server.port", "8012")
  // db configuration flags
  cmdline.AddOption("", "database.host", "HOST", "host of database")
  cmdline.SetOptionDefault("database.host", "localhost")
  cmdline.AddOption("", "database.port", "PORT", "port of database")
  cmdline.SetOptionDefault("database.port", "5432")
  cmdline.AddOption("", "database.name", "NAME", "name of database")
  cmdline.SetOptionDefault("database.name", "postgres")
  cmdline.AddOption("", "database.username", "USERNAME", "username for database user")
  cmdline.SetOptionDefault("database.username", "postgres")
  cmdline.AddOption("", "database.password", "PASSWORD", "password for the database user")
  cmdline.SetOptionDefault("database.password", "postgres")

  cmdline.AddFlag("v", "verbose", "log more information")
  cmdline.Parse(os.Args)

  // application general configuration
  a.Config.Language = cmdline.OptionValue("language")

  // a server configuration values
  a.Config.Server.Host = cmdline.OptionValue("server.host")
  a.Config.Server.Port = cmdline.OptionValue("server.port")

  // database configuration value
  a.Config.Database.Host     = cmdline.OptionValue("database.host")
  a.Config.Database.Port     = cmdline.OptionValue("database.port")
  a.Config.Database.Name     = cmdline.OptionValue("database.name")
  a.Config.Database.Username = cmdline.OptionValue("database.username")
  a.Config.Database.Password = cmdline.OptionValue("database.password")
}

var BuildTime = "unset"
var Commit 		= "unset"
var Release 	= "unset"

func registerVersion(a *internal.Application) {
  a.Version = &internal.Version{BuildTime, Commit, Release}
}

func main() {
  a := internal.Application{}

  registerCliConfiguration(&a)
  registerVersion(&a)

  a.Initialize()
  a.Run()
}
