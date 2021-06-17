package resource_test

import(
	bytes  "bytes"
	colors   "github.com/cucumber/godog/colors"
	flag     "github.com/spf13/pflag"
	fmt 		 "fmt"
	godog    "github.com/cucumber/godog"
	http 		 "net/http"
	internal "github.com/gpenaud/needys-api-resource/internal"
	os 			 "os"
	testing  "testing"
	"encoding/json"
)

var application internal.Application

func init() {
	godog.BindCommandLineFlags("godog.", &opts)

	application.Version = &internal.Version{}
	application.Config = &internal.Configuration{}

	application.Config.Verbosity 	 = "fatal"
	application.Config.Environment = "development"
	application.Config.LogFormat 	 = "text"
	application.Config.Server.Host = "0.0.0.0"
	application.Config.Server.Port = "8012"
}

var opts = godog.Options{
  Output: colors.Colored(os.Stdout),
  Format: "progress", // can define default values
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^I send "([^"]*)" request to "([^"]*)"$`, iSendRequestTo)
	ctx.Step(`^the response code should be (\d+)$`, theResponseCodeShouldBe)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = flag.Args()

	status := godog.TestSuite{
		Name: "godogs",
		// TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer: InitializeScenario,
		Options: &opts,
	}.Run()

	os.Exit(status)
}

var res *http.Response

func iSendRequestTo(method, endpoint string) error {
	client := &http.Client{}

	server := fmt.Sprintf("http://%s:%s",
		application.Config.Server.Host,
		application.Config.Server.Port,
	)

	var data map[string]string

	switch endpoint {
	case "/POST":
		data = map[string]string{
			"type": "outdoor",
			"description": "Du bon gros sexe des familles !",
		}
	case "/PUT":
		data = map[string]string{
			"id": "3",
			"description": "Du bon gros sexe, mais sans la famille cette fois !",
		}
	default:
		data = map[string]string{}
	}

	payload, err := json.Marshal(data)

  req, err := http.NewRequest(method, server+endpoint, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	if err != nil {
		return fmt.Errorf("could not create request %s", err.Error())
	}

 	res, err = client.Do(req)
  if err != nil {
    return fmt.Errorf("could not send request %s", err.Error())
  }

	// TODO find a soltion to clean database between test calls
	// if (endpoint != "initialize_db") {
	// 	iSendRequestTo("GET", "initialize_db")
	// }

	return nil
}

func theResponseCodeShouldBe(code int) error {
	if code != res.StatusCode {
		return fmt.Errorf("expected response code to be: %d, but actual is: %d", code, res.StatusCode)
	}

	return nil
}
