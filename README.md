# cf-accounting-report
Pulls Accounting reports from Pivotal Cloud Foundry

## Installing

Install by 

```bash
cf install-plugin <insert url here>
```

Download Urls can be found on the [releases page](https://github.com/mcb/cf-accounting-report/releases) for each platform. Right-Click and copy target link.

## Usage

```
% cf accounting-report
+------+------+-------+-----------+---------+----------+
| TYPE | YEAR | MONTH |  AVERAGE  | MAXIMUM |  HOURS   |
+------+------+-------+-----------+---------+----------+
| AI   | 2018 | all   |       0.1 |       7 | 787.0539 |
+------+------+-------+-----------+---------+----------+
| AI   | 2018 |     8 | 1.0711303 |       7 | 787.0539 |
+------+------+-------+-----------+---------+----------+
Report Date: 2018-08-31 14:47:17 UTC
```

This will list usage data for all application instances, service instances can be displayed using the ```--services``` switch

```
% cf accounting-report --services
+---------+----------------+------+-------+---------+---------+------------+
|  TYPE   |      NAME      | YEAR | MONTH | AVERAGE | MAXIMUM |   HOURS    |
+---------+----------------+------+-------+---------+---------+------------+
| Service | p-rabbitmq     | 2018 |     1 |       0 |       0 |          0 |
+---------+----------------+------+-------+---------+---------+------------+
| Service | p-rabbitmq     | 2018 |     2 |       0 |       0 |          0 |
+---------+----------------+------+-------+---------+---------+------------+
| Service | p-rabbitmq     | 2018 |     3 |       0 |       0 |          0 |
+---------+----------------+------+-------+---------+---------+------------+
| Service | p-rabbitmq     | 2018 |     4 |       0 |       0 |          0 |
+---------+----------------+------+-------+---------+---------+------------+
Report Date: 2018-08-31 14:34:32 UTC

```

If you want pure JSON Data, this can be enabled using the ```--output-json``` switch.

```bash
% cf accounting-report --output-json
{"report_time":"2018-08-31 14:47:55 UTC","monthly_reports":[{"year":2018,"month":8,"average_app_instances":1.0711725,"maximum_app_instances":7,"app_instance_hours":787.0961}],"yearly_reports":[{"year":2018,"month":0,"average_app_instances":0.1,"maximum_app_instances":7,"app_instance_hours":787.0961}]}
```


## Development

```bash
go install ./cmd/accounting-report && \
    cf install-plugin $GOPATH/bin/accounting-report -f && \
    cf accounting-report
```


## Credits

Credits to [govau](https://github.com/govau) for their implementation of the apiClient that I used for this project.
