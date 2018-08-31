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

type servicePlan struct {
  Name    string      `json:"service_name"`
  Usage   []service   `json:"usages"`
}


// Why oh why does it have a different format than the service Plans?!?
type servicePlanYearly struct {
  Name      string    `json:"service_name"`
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
  table.SetCaption(true, "Report Date: "+appReport.ReportTime)
  table.Render()

  return nil
}

// Retrieves Services Usage and prints it into table or json

func (c *accountingReport) GetServiceUsage(client *apiClient, out io.Writer, outputJSON bool) error {

  var serviceReport struct {
      ReportTime  string                `json:"report_time"`
      Monthly     []servicePlan         `json:"monthly_service_reports"`
      Yearly      []servicePlanYearly   `json:"yearly_service_report"`
    }

  err := client.Get("/system_report/service_usages", &serviceReport)
  
  if err != nil {
    return err
  }

  if outputJSON {
    return json.NewEncoder(out).Encode(serviceReport)
  }

  table := tablewriter.NewWriter(out)
  table.SetRowLine(true)
  table.SetHeader([]string{"Type", "Name", "Year", "Month", "Average", "Maximum", "Hours"})

  for _, monthly := range serviceReport.Monthly {
    var name = monthly.Name
    for _, usage := range monthly.Usage {
      table.Append([]string{"Service", name, fmt.Sprint(usage.Year), fmt.Sprint(usage.Month), fmt.Sprint(usage.Average), fmt.Sprint(usage.Maximum), fmt.Sprint(usage.Hours)})
    }
  }
  for _, yearly := range serviceReport.Yearly {
    table.Append([]string{"Service", yearly.Name, fmt.Sprint(yearly.Year), "all", fmt.Sprint(yearly.Average), fmt.Sprint(yearly.Maximum), fmt.Sprint(yearly.Hours)})
  }
  table.SetCaption(true, "Report Date: "+serviceReport.ReportTime)
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
  applications := true
  services := false

  fs := flag.NewFlagSet("accounting-report", flag.ExitOnError)
  fs.BoolVar(&outputJSON, "output-json", false, "if set sends JSON to stdout instead of table rendering")
  fs.BoolVar(&applications, "applications", false, "if set only prints applications data")
  fs.BoolVar(&services, "services", false, "if set only prints services data")

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
    var err error
    if services {
      err = c.GetServiceUsage(client, os.Stdout, outputJSON)
    } else {
      err = c.GetAppUsage(client, os.Stdout, outputJSON)
    }

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
      Build: 2,
    },
    MinCliVersion: plugin.VersionType{
      Major: 6,
      Minor: 7,
      Build: 0,
    },
    Commands: []plugin.Command{
      {
        Name:     "accounting-report",
        HelpText: "lists usage data of purchased resources with applications being default",

        // UsageDetails is optional
        // It is used to show help of usage of each command
        UsageDetails: plugin.Usage{
          Usage: "cf accounting-report",
          Options: map[string]string{
            "output-json": "if set prints JSON to stdout instead of a rendered table",
            "applications": "if set only prints applications data",
            "services": "if set only prints services data",
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
