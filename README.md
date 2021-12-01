# rpki_exporter
This prometheus exporter will export rpki status for specific prefix and AS combination. Status can be set as:
  - invalid
  - valid
  - not-found

This exporter uses `https://rpki-validator.ripe.net/` to check the status of rpki.

## Install
```go install github.com/tomvil/rpki_exporter@latest```

## Configuration file
Configuration example:
```yaml
refresh_interval: 3600 # not required, 600 is the default value. It's here just to show that it can be changed if needed.
targets:
  - as: 15169
    prefixes: # All prefixes must be a valid network address with a prefix at the end!
      - 2001:4860:4864::/48
      - 2404:6800:4001::/48
      - 8.8.4.0/24
      - 8.8.8.0/24
```

- This example can be found in `config.example.yaml` file.
- You can have multiple AS numbers defined and of course multiple prefixes per AS.

## Usage
1. Generate configuration file.
2. `./rpki_exporter --config.file your-configuration-file.yaml`
    - If `--config.file` flag is not set, by default exporter will try to open `config.yaml` file. 

## Flags
Name     | Description | Default
---------|-------------|---------
web.listen-address | Address on which to expose metrics and web interface. | :9959
web.metrics-path | Path under which to expose metrics. | /metrics
config.file | Path to config file | config.yaml

## Metrics
```
# HELP rpki_queries_failed_total Number of failed queries
# TYPE rpki_queries_failed_total counter
rpki_queries_failed_total 0
# HELP rpki_queries_success_total Number of successful queries
# TYPE rpki_queries_success_total counter
rpki_queries_success_total 4
# HELP rpki_status RPKI Status of the prefix (0 - invalid, 1 - valid, 2 - not found)
# TYPE rpki_status gauge
rpki_status{asn="AS15169",prefix="2001:4860:4864::/48"} 1
rpki_status{asn="AS15169",prefix="2404:6800:4001::/48"} 1
rpki_status{asn="AS15169",prefix="8.8.4.0/24"} 1
rpki_status{asn="AS15169",prefix="8.8.8.0/24"} 1
```

