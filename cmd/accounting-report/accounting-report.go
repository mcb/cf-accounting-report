package main

import (
  "fmt"
  "flag"
  "os"
  "log"
  "io"
  "strings"
  "net/http"
  "encoding/json"
  "errors"

  "github.com/olekukonko/tablewriter"
  "code.cloudfoundry.org/cli/plugin"
)


type accountingReport struct{}

type apiClient struct {
  // API url, ie "https://api.system.example.com"
  API string

  // Authorization header, ie "bearer eyXXXXX"
  Authorization string
}

type application struct {
  Year      int       `json:"year"`
  Month     int       `json:"month"`
  Average   float32   `json:"average_app_instances"`
  Maximum   float32   `json:"maximum_app_instances"`
  Hours     float32   `json:"app_instance_hours"`
}

type service struct {
  Year      int       `json:"year"`
  Month     int       `json:"month"`
  Average   float32   `json:"average_instances"`
  Maximum   float32   `json:"maximum_instances"`
  Hours     float32   `json:"duration_in_hours"`
}

func (c *accountingReport) GetAppUsage(client *apiClient, out io.Writer, outputJSON bool) error {

  var appReport struct {
      ReportTime  string          `json:"report_time"`
      Monthly     []application   `json:"monthly_reports"`
      Yearly      []application   `json:"yearly_reports"`
    }

  err := client.Get("/system_report/app_usages", &appReport)
  
  if err != nil {
    return err
  }

  if outputJSON {
    return json.NewEncoder(out).Encode(appReport)
  }

  table := tablewriter.NewWriter(out)
  table.SetRowLine(true)
  table.SetHeader([]string{"Type", "Year", "Month", "Average", "Maximum", "Hours"})

  for _, yearly := range appReport.Yearly {
    table.Append([]string{"AI", fmt.Sprint(yearly.Year), "all", fmt.Sprint(yearly.Average), fmt.Sprint(yearly.Maximum), fmt.Sprint(yearly.Hours)})
  }
  
  for _, monthly := range appReport.Monthly {
    table.Append([]string{"AI", fmt.Sprint(monthly.Year), fmt.Sprint(monthly.Month), fmt.Sprint(monthly.Average), fmt.Sprint(monthly.Maximum), fmt.Sprint(monthly.Hours)})
  }
  table.Render()

  return nil
}


// Get makes a GET request, where r is the relative path, and rv is json.Unmarshalled to
func (ac *apiClient) Get(r string, rv interface{}) error {
  req, err := http.NewRequest(http.MethodGet, ac.API+r, nil)
  if err != nil {
    return err
  }
  req.Header.Set("Authorization", ac.Authorization)
  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return errors.New("bad status code")
  }

  return json.NewDecoder(resp.Body).Decode(rv)
}


func newApiClient(cliConnection plugin.CliConnection) (*apiClient, error) {
 at, err := cliConnection.AccessToken()
  if err != nil {
    return nil, err
  }

  api, err := cliConnection.ApiEndpoint()
  if err != nil {
    return nil, err
  }

  appUsageApi := strings.Replace(api, "api", "app-usage", 1)

  return &apiClient{
    API:           appUsageApi,
    Authorization: at,
  }, nil
}


// Run must be implemented by any plugin because it is part of the
// plugin interface defined by the core CLI.
func (c *accountingReport) Run(cliConnection plugin.CliConnection, args []string) {
  outputJSON := false

  fs := flag.NewFlagSet("accounting-report", flag.ExitOnError)
  fs.BoolVar(&outputJSON, "output-json", false, "if set sends JSON to stdout instead of table rendering")
  err := fs.Parse(args[1:])
  if err != nil {
    log.Fatal(err)
  }

  client, err := newApiClient(cliConnection)
  if err != nil {
    log.Fatal(err)
  }

  switch args[0] {
    case "accounting-report":
    err := c.GetAppUsage(client, os.Stdout, outputJSON)
    if err != nil {
      log.Fatal(err)
    }
  }
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
func (c *accountingReport) GetMetadata() plugin.PluginMetadata {
  return plugin.PluginMetadata{
    Name: "accounting-report",
    Version: plugin.VersionType{
      Major: 0,
      Minor: 0,
      Build: 1,
    },
    MinCliVersion: plugin.VersionType{
      Major: 6,
      Minor: 7,
      Build: 0,
    },
    Commands: []plugin.Command{
      {
        Name:     "accounting-report",
        HelpText: "lists usage data of purchased resources",

        // UsageDetails is optional
        // It is used to show help of usage of each command
        UsageDetails: plugin.Usage{
          Usage: "cf accounting-report",
          Options: map[string]string{
            "output-json": "if set prints JSON to stdout instead of a rendered table",
            "type":        "if set only prints set type (not implemented yet)",
          },
        },
      },
    },
  }
}

// Unlike most Go programs, the `Main()` function will not be used to run all of the
// commands provided in your plugin. Main will be used to initialize the plugin
// process, as well as any dependencies you might require for your
// plugin.
func main() {
  plugin.Start(new(accountingReport))
}
