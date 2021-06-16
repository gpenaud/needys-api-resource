package internal

func (a *Application) isDatabaseReachable() (err error) {
  _, err = a.DB.Query("SELECT null")
  return err
}
