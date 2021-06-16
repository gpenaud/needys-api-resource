package internal

func (a *Application) I18n() (I18n) {
  return a.I18nWithEnforcedLanguage(a.Config.Language)
}

func (a *Application) I18nWithEnforcedLanguage(language string) (I18n) {
  switch language {
  case FR:
    return I18n{
      DatabaseNotAvailable:     "La base de  donnée n'est pas disponible",
      DatabaseNotInitializable: "La base de  donnée n'est pas initialisable",
      ResourceNotfound:         "La ressource n'a pas été trouvée",
      InvalidPayloadRequest:    "La charge utile est invalide",
      InvalidResourceID:        "L'ID de la ressource est invalide",
    }
  default:
    return I18n{
      //messages
      StartingApplicationMessage:      "Starting needys-api-resource on %s:%s...\n > build time: %s\n > release: %s\n > commit: %s\n",
      // errors
      DatabaseNotAvailable:            "Database is not available",
      DatabaseNotInitializable:        "Database is not initializable",
      ResourceNotfound:                "The resource is not found",
      InvalidPayloadRequest:           "The payload is invalid",
      InvalidResourceID:               "The resource ID is invalid",
      // units test erros
      UnxpectedResponseCode:           "Expected response code %d. Got %d\n",
      UnexpectedNonEmptyArray:         "Expected an empty array. Got %s",
      UnexpectedIDNotRemainTheSame:    "Expected the id to remain the same (%v). Got %v",
      UnexpectedTypeNotChanged:        "Expected the type to change from '%v' to '%v'. Got '%v'",
      UnexpectedDescriptionNotChanged: "Expected the description to change from '%v' to '%v'. Got '%v'",
    }
  }
}

type I18n struct {
  //messages
  StartingApplicationMessage      string
  // errors
  DatabaseNotAvailable            string
  DatabaseNotInitializable        string
  ResourceNotfound                string
  InvalidPayloadRequest           string
  InvalidResourceID               string
  // units test erros
  UnxpectedResponseCode           string
  UnexpectedNonEmptyArray         string
  UnexpectedIDNotRemainTheSame    string
  UnexpectedTypeNotChanged        string
  UnexpectedDescriptionNotChanged string
}

const (
  EN string = "EN"
  FR string = "FR"
)
