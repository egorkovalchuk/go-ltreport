# Load Test Report (go-ltreport)
A Go-based utility designed to automate load testing results analysis and summary report generation.

This tool helps developers, QA engineers, and DevOps specialists quickly transform raw metrics from various data sources into structured PDF reports, or automatically publish them directly to Confluence pages and Allure TestOps.

## 🚀 Key Features
- **Multi-Platform Metrics Collection:** Native integration with InfluxDB (JMeter), Grafana, ClickHouse, Prometheus, Graphite, Telegraf, and HP Service Manager (HPSM/FSM).
- **Smart Threshold Analysis:** Automatic metrics analysis against predefined threshold values for the entire test scope as well as individual scenarios/transactions. Supports conditional comparison operators (`>` and `<`).
- **Visual PDF Reports:** Generates comprehensive documents containing interactive tables, incident logs, and embedded Grafana panel charts (screenshots).
- **Grafana Templating via CSV:** Supports bulk, parallel chart generation using template layouts driven by an input CSV configuration file. Optional `ThresholdFilter` to show graphs only when thresholds are violated.
- **Confluence Automation:** Automatically uploads generated reports to target Confluence directories as child pages with the PDF report attached. Supports both password and token-based authentication.
- **Allure Framework Integration:** Exports individual threshold check results to Allure as test cases mapped to `passed`, `failed`, or `broken` statuses with attachments.
- **Multi-Encoding Support:** Automatic detection and conversion of Windows-1251, KOI8-R, ISO-8859-5, and UTF-8 encodings.
- **External Config Files:** Supports loading JMeter, Grafana Dashboards, and Grafana Template configurations from separate JSON files via `JmeterConfig`, `GrafanadashConfig`, and `GrafanadashTemplateConfig` settings.

## 📐 Project & Codebase Architecture
The project codebase is organized following idiomatic **Standard Go Project Layout** patterns:

* **`cmd/ltreport/`** — Application entry point. Contains `go-ltreport.go` (main CLI handler, report orchestration) and `init.go` (variable initialization, time processing, temp file management).
* **`internal/`** — Private application code directory.
  * `internal/reportdata/` — Core data structures, configuration loading, validation, threshold parsing, and utility functions.
  * `internal/source/` — Integration adapters and clients for metric sources.
    * `internal/source/influx/` — InfluxDB (JMeter) client.
    * `internal/source/grafana/` — Grafana dashboard and alert client.
    * `internal/source/prometheus/` — Prometheus metrics client.
    * `internal/source/graphite/` — Graphite API client.
    * `internal/source/telegraf/` — Telegraf data source client.
    * `internal/source/clickhouse/` — ClickHouse SQL client.
  * `internal/pdf/` — PDF report rendering engine for document formatting, text layout, and chart attachment.
  * `internal/allure/` — Client implementation for sending test execution results to Allure.
  * `internal/confluence/` — Confluence API client for page creation and PDF attachment.
  * `internal/hpsm/` — Integration client for fetching incidents from HP Service Manager (FSM).
  * `internal/logger/` — Structured logging with file and optional console output.
  * `internal/notify/` — Notification module.
  * `internal/fonts/` — Embedded font files (Times New Roman variants) for PDF rendering.
* **`example/`** — Configuration templates and mock data sets to help you get started quickly.

## 🔄 Workflow Pipeline
Upon execution, the utility runs a sequential data collection and processing pipeline:

```
[Start main()] ──> [Parse CLI Flags] ──> [Read config.json]
│
└──> [Validate Dates & Create tmp/]
│
└──> [Execute StartReport()]
│
├──> [Initialize PDF Engine & Allure Client]
├──> [Parallel Metrics Collection: InfluxDB/JMeter, Grafana, ClickHouse, HPSM]
├──> [Threshold Violations Analysis -> Log Incidents]
├──> [Render PDF Document (Tables + Dashboard Screenshots)]
└──> [Upload PDF to Confluence (Optional) + Push Archive to Allure]
│
└──> [Cleanup tmp/ (via -rm flag)] ──> [Send Notifications (if not -ban)] ──> [Exit]
```

---

## ⚙️ Command Line Interface Flags

The following parameters are available when running the application binary:

| Flag | Data Type | Default Value | Description |
| :--- | :--- | :--- | :--- |
| `-config` | string | `config.json` | Path to a custom application configuration file. |
| `-d` | boolean | `false` | Enables debug mode for verbose logging output. |
| `-v` | boolean | `false` | Prints the utility's current version (Current version: `0.6.0.2`). |
| `-h` | boolean | `false` | Displays help message with available flags. |
| `-c` | boolean | `false` | Enables console output in addition to log file. |
| `-hour` | boolean | `false` | Automatically generates a report covering the trailing hour. |
| `-start` | string | `""` | Report window start time using the `2006.01.02 15:04` format (e.g., `"2026.01.31 15:00"`). |
| `-end` | string | `""` | Report window end time using the `2006.01.02 15:04` format (e.g., `"2026.01.31 18:00"`). |
| `-rm` | boolean | `false` | Automatically removes temporary PNG files after report generation. |
| `-ban` | boolean | `false` | Disables sending notifications after report completion. |
| `-fsmlogin` | string | `""` | Overrides FSM/HP Service Manager login from config file. |
| `-fsmpass` | string | `""` | Overrides FSM/HP Service Manager password from config file. |
| `-conflproxy` | string | `""` | Sets Confluence proxy URL (format: `http://user:password@host:port`). |
| `-ConfToken` | string | `""` | Overrides Confluence authentication token from config file. |
| `-CHUser` | string | `""` | Overrides ClickHouse username from config file. |
| `-CHPassword` | string | `""` | Overrides ClickHouse password from config file. |

> **Note:** CLI flags such as `-fsmlogin`, `-fsmpass`, `-conflproxy`, `-ConfToken`, `-CHUser`, and `-CHPassword` take precedence over values specified in the configuration file.

---

## 📦 Configuration Specification (`config.json`)

The application behavior is fully customized via a JSON configuration file. The real configuration structure uses **PascalCase** keys. Below is the actual structure matched to the codebase:

```json
{
  "ReportFilename": "LTReport",
  "ReportMask": "02012006",
  "ReportPath": "/tmp/",
  "Confluence": {
    "ReportConfluenceOn": true,
    "ReportConfluenceURL": "https://confluence.company.com",
    "ReportConfluenceId": "12345678",
    "ReportConfluenceSpace": "SPACE",
    "ReportConfluenceLogin": "admin",
    "ReportConfluencePass": "password",
    "ReportConfluenceToken": "your-confluence-token",
    "ReportConfluenceProxy": ""
  },
  "Allure": {
    "ReportAllure": true,
    "ReportEndPoint": "http://allure-server:8080",
    "Token": "your-allure-token",
    "ReportPath": "./allure-results/",
    "ReportAllureProxy": ""
  },
  "ReportIM": {
    "Token": "",
    "LoginFSM": "fsm-user",
    "PassFSM": "fsm-password",
    "FSMConnect": "fsm.company.com",
    "FSMTU": "TU-001"
  },
  "ClickHouse": {
    "Server": "localhost:8123",
    "User": "default",
    "Pass": "",
    "Token": "",
    "Query": [
      {
        "sql": "SELECT * FROM lt_metrics WHERE date = today()",
        "dbname": "analytics",
        "name": "Daily Metrics Summary"
      }
    ]
  },
  "ReportOn": {
    "ReportJmeter": true,
    "ReportIM": true,
    "ReportScenario": true,
    "ReportDash": true,
    "ClickHouse": true
  },
  "JmeterConfig": "",
  "GrafanadashConfig": [],
  "GrafanadashTemplateConfig": [],
  "Jmeter": {
    "JmeterInflux": "http://localhost:8086",
    "JmeterLoginInflux": "",
    "JmeterPassInflux": "",
    "JmeterQuery": "SELECT mean(\"rate\"), mean(\"avg\"), mean(\"errpct\"), percentile(\"avg\", 95), max(\"avg\"), percentile(\"avg\", 50) FROM \"delta\" WHERE ",
    "JmeterQueryGroup": "GROUP BY \"suite\", time(1d) fill(0)",
    "JmeterQueryField": [
      {"Name": "Rate", "Description": "Average rate %d ops"},
      {"Name": "Avg", "Description": "Average latency %d ms"},
      {"Name": "ErrorPercent", "Description": "Error %d %%" },
      {"Name": "Percentile", "Description": "Percentile latency %d ms"},
      {"Name": "Maxlatency", "Description": "Max latency %d ms"},
      {"Name": "Percentile50", "Description": "Percentile 50 %d ms"}
    ],
    "JmeterQueryThreshold": [
      {"Name": "oapi-get", "ErrorField": "ErrorPercent", "Threshold": 3, "Description": "Error rate exceeded %d in test %s current %d"}
    ],
    "JmeterQueryScenario": "SELECT max(\"pct90.0\"), max(\"max\"), mean(\"count\") FROM \"details\" WHERE ",
    "JmeterQueryScnrGroup": " GROUP BY time(1d), \"application\", \"transaction\", \"statut\" fill(0) ORDER BY time DESC",
    "JmeterQueryScnrField": [
      {"Name": "Percentile90", "Description": "Percentile90 latency %d ms"},
      {"Name": "Maxlatency", "Description": "Max latency %d ms"},
      {"Name": "Rate", "Description": "Average rate %d ops"}
    ],
    "JmeterQueryScnrThreshold": [
      {"Name": "oapi_get", "NameThread": "getToken", "ErrorField": "Percentile90", "Threshold": 5, "Statut": "ok", "Description": "High latency %d ms in scenario %s current %d ms"}
    ]
  },
  "GrafanaDash": [
    {
      "Name": "ESB CPU",
      "UrlDash": "https://grafana.company.com/d/000000879/wf?orgId=1&refresh=5m",
      "AuthHeader": "Bearer token-value",
      "UrlPanel": "",
      "UrlImg": "https://grafana.company.com/render/d-solo/000000879/wf?height=500&orgId=1&panelId=88",
      "Query": "SELECT percentile(\"usage_active\",95) FROM \"ret_30w\".\"cpu\" WHERE (\"host\" =~ /server_name/) ",
      "SourceType": 1,
      "UrlQuery": "https://grafana.company.com/api/datasources/proxy/459/query?db=telegraf&q=",
      "UrlQueryGroup": "GROUP BY time(1d) fill(none)",
      "Threshold": 90,
      "ThDescription": "CPU threshold exceeded",
      "Tag": "Product:Test;layer:system",
      "Size": {"Width": 120, "Height": 50},
      "ThresholdFilter": false,
      "FileList": "",
      "FilePettern": ""
    }
  ],
  "GrafanadashTemplate": [
    {
      "Name": "ESB CPU Template",
      "UrlDash": "https://grafana.company.com/d/000000879/wf?orgId=1&refresh=5m",
      "AuthHeader": "Bearer token-value",
      "UrlPanel": "",
      "UrlImg": "https://grafana.company.com/render/d-solo/000000879/wf?height=500&orgId=1&panelId=88",
      "Query": "SELECT percentile(\"usage_active\",95) FROM \"ret_30w\".\"cpu\" WHERE (\"host\" =~ /server_name/) ",
      "SourceType": 1,
      "UrlQuery": "https://grafana.company.com/api/datasources/proxy/459/query?db=telegraf&q=",
      "UrlQueryGroup": "GROUP BY time(1d) fill(none)",
      "Threshold": 90,
      "ThDescription": "CPU threshold exceeded",
      "Tag": "Product:Test",
      "Size": {"Width": 120, "Height": 50},
      "ThresholdFilter": true,
      "FileList": "templates.csv",
      "FilePettern": "server;name"
    }
  ]
}
```

> **Note:** The configuration supports external files for JMeter, Grafana Dashboards, and Grafana Templates via `JmeterConfig`, `GrafanadashConfig`, and `GrafanadashTemplateConfig` fields respectively. These allow keeping configurations modular.

### Grafana Metric Validation Rules:
The engine applies two distinct evaluation strategies for Grafana dashboard targets:
1. **Conditional Expressions (`Comparison`):** Checks whether the source database percentile values violate the `Threshold` setting based on conditional operators (`>` or `<`). The `Threshold` field can be either a numeric value (defaults to `>` comparison) or a string like `"<10"` or `">90"` to specify the comparison operator explicitly.
2. **Built-in Dashboard Alerts (`AlertID`):** When given an alert ID, the engine tracks the state history in Grafana and registers a critical incident if an active alert is captured.

### Source Type Reference for Grafana Dashboards:
| SourceType | Data Source |
|:---:|---|
| `1` | Telegraf / InfluxDB |
| `2` | Prometheus |
| `3` | Graphite |

---

## 🛠 Compilation and Build Targets

Building the application binary requires **Go 1.24+** and the `make` utility installed on your local workstation.

* **Build the application binary:**
  ```bash
  make build
  ```
* **Run test suites:**
  ```bash
  make test
  ```
* **Clean up compiled artifacts:**
  ```bash
  make clean
  ```

---

## 📖 Usage Examples

### 1. Generate a default report (automatically defaults to reading `config.json`)
```bash
./ltreport
```

### 2. Generate a custom-scoped report using an absolute time frame and specific configuration profile
```bash
./ltreport -start "2026.12.31 09:00" -end "2026.12.31 18:00" -config custom_config.json
```

### 3. Run execution in debug mode and automatically flush temporary files on finish
```bash
./ltreport -d -rm -config config.json
```

---

## 🤝 Confluence Publishing Integration

Enabling the `ReportConfluenceOn` option triggers the following publication cycle:
1. Creates a nested child page inside your Confluence space structure beneath the designated `ReportConfluenceId`.
2. Formats markdown contents layout onto the new page body and attaches the master PDF binary report.
3. Generates a live publication tracking URL, which is automatically synchronized and appended to your Allure test runs.

Authentication supports both **password-based** (`ReportConfluenceLogin` / `ReportConfluencePass`) and **token-based** (`ReportConfluenceToken`) methods. A proxy can be specified via `ReportConfluenceProxy`.

---

## 👨‍💻 Contribution Guidelines

We highly appreciate community contributions to expand this toolset!
- Spotted a bug or want to submit an improvement proposal? Open an [Issue](https://github.com).
- Ready to fix an issue or submit a feature extension? Open a [Pull Request](https://github.com).

---

## 📄 License

This software is distributed under the terms of the **MIT License**.

**Author:** [Egor Kovalchuk](https://github.com)  
**Current Version:** `0.6.0.2`