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

// BasicPlugin is the struct implementing the interface defined by the core CLI. It can
// be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type accountingReport struct{}

// simpleClient is a simple CloudFoundry client
type apiClient struct {
  // API url, ie "https://api.system.example.com"
  API string

  // Authorization header, ie "bearer eyXXXXX"
  Authorization string
}

type report struct {
  Year      int   `json:"year"`
  Month     int   `json:"month"`
  Average   float32   `json:"average_app_instances"`
  Maximum   float32   `json:"maximum_app_instances"`
  Hours     float32   `json:"app_instance_hours"`
}


func (c *accountingReport) GetAccountUsage(client *apiClient, out io.Writer, outputJSON bool) error {
  var res struct {
      ReportTime  string     `json:"report_time"`
      Monthly     []report   `json:"monthly_reports"`
      Yearly      []report   `json:"yearly_reports"`
    }

  err := client.Get("/system_report/app_usages", &res)
  
  if err != nil {
    return err
  }

  if outputJSON {
    return json.NewEncoder(out).Encode(res)
  }

  table := tablewriter.NewWriter(out)
  table.SetRowLine(true)
  table.SetHeader([]string{"Type", "Year", "Month", "Average", "Maximum", "Hours"})

  for _, yearly := range res.Yearly {
    table.Append([]string{"AI", fmt.Sprint(yearly.Year), "all", fmt.Sprint(yearly.Average), fmt.Sprint(yearly.Maximum), fmt.Sprint(yearly.Hours)})
  }
  
  for _, monthly := range res.Monthly {
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
//
// Run(....) is the entry point when the core CLI is invoking a command defined
// by a plugin. The first parameter, plugin.CliConnection, is a struct that can
// be used to invoke cli commands. The second paramter, args, is a slice of
// strings. args[0] will be the name of the command, and will be followed by
// any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing
// user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
// 1 should the plugin exits nonzero.
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
    err := c.GetAccountUsage(client, os.Stdout, outputJSON)
    if err != nil {
      log.Fatal(err)
    }
  }
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as seperate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf basic-plugin-command` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.
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
          Usage: "basic-plugin-command\n   cf basic-plugin-command",
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
  // Any initialization for your plugin can be handled here
  //
  // Note: to run the plugin.Start method, we pass in a pointer to the struct
  // implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
  //
  // Note: The plugin's main() method is invoked at install time to collect
  // metadata. The plugin will exit 0 and the Run([]string) method will not be
  // invoked.
  plugin.Start(new(accountingReport))
  // Plugin code should be written in the Run([]string) method,
  // ensuring the plugin environment is bootstrapped.
}
