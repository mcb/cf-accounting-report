# cf-accounting-report
Pulls Accounting reports from Pivotal Cloud Foundry

## Usage

```bash
cf accounting-report
```

## Development

```bash
go install ./cmd/accounting-report && \
    cf install-plugin $GOPATH/bin/accounting-report -f && \
    cf accounting-report
```
