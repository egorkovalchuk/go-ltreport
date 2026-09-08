Here is the complete, professional English version of your documentation. You can use this text to update your main README.md file while keeping the Russian text for readme.ru.md.
------------------------------

# Load Test Report (go-ltreport)
A Go-based utility designed to automate load testing results analysis and summary report generation.

This tool helps developers, QA engineers, and DevOps specialists quickly transform raw metrics from various data sources into structured PDF reports, or automatically publish them directly to Confluence pages and Allure TestOps.

---## 🚀 Key Features
- **Multi-Platform Metrics Collection:** Native integration with InfluxDB (JMeter), Grafana, ClickHouse, Prometheus, Graphite, and HP Service Manager (HPSM).
- **Smart Threshold Analysis:** Automatic metrics analysis against predefined threshold values for the entire test scope as well as individual scenarios/transactions.
- **Visual PDF Reports:** Generates comprehensive documents containing interactive tables, incident logs, and embedded Grafana panel charts (screenshots).
- **Grafana Templating via CSV:** Supports bulk, parallel chart generation using template layouts driven by an input CSV configuration file.
- **Confluence Automation:** Automatically uploads generated reports to target Confluence directories as child pages with the PDF report attached.
- **Allure Framework Integration:** Exports individual threshold check results to Allure as test cases mapped to `passed`, `failed`, or `broken` statuses.

---## 📐 Project & Codebase Architecture
The project codebase is organized following idiomatic **Standard Go Project Layout** patterns:

* **`cmd/ltreport/`** — Application entry point. Contains `go-ltreport.go`, which handles CLI flags parsing, logging/configuration initialization, and fires up the main application engine.
* **`internal/`** — Private application code directory.
  * `internal/config/` — Configuration file loading and validation logic.
  * `internal/source/` — Integration adapters and clients for metric sources (InfluxDB, ClickHouse, Prometheus, Graphite, Grafana).
  * `internal/pdf/` — PDF report rendering engine for document formatting, text layout, and chart attachment.
  * `internal/allure/` — Client implementation for sending test execution results to Allure.
  * `internal/hpsm/` — Integration client for fetching incidents from HP Service Manager.
* **`example/`** — Configuration templates and mock data sets to help you get started quickly.

---## 🔄 Workflow Pipeline
Upon execution, the utility runs a sequential data collection and processing pipeline:


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


---

## ⚙️ Command Line Interface Flags

The following parameters are available when running the application binary:

| Flag | Data Type | Default Value | Description |
| :--- | :--- | :--- | :--- |
| `-config` | string | `config.json` | Path to a custom application configuration file. |
| `-d` | boolean | `false` | Enables debug mode for verbose logging output. |
| `-v` | boolean | `false` | Prints the utility's current version (Current version: `0.6.0.2`). |
| `-hour` | boolean | `false` | Automatically generates a report covering the trailing hour. |
| `-start` | string | `""` | Report window start time using the `YYYY.MM.DD HH:MM` format (e.g., `"2026.01.31 15:00"`). |
| `-end` | string | `""` | Report window end time using the `YYYY.MM.DD HH:MM` format. |
| `-rm` | boolean | `false` | Automatically deletes temporary assets (the `tmp/` folder) after a successful run. |
| `-ban` | boolean | `false` | Disables sending out report completion notifications. |
| `-CHUser` | string | `""` | Overrides the ClickHouse username specified in the config file. |
| `-CHPassword`| string | `""` | Overrides the ClickHouse password specified in the config file. |
| `-ConfToken` | string | `""` | Overrides the Confluence API access token. |
| `-conflproxy`| string | `""` | Specifies a proxy server for Confluence connections (`http://user:password@url:port`). |
| `-fsmlogin` | string | `""` | User credentials login for HP Service Manager (FSM) integration. |
| `-fsmpass` | string | `""` | User credentials password for HP Service Manager (FSM) integration. |
| `-h` | boolean | `false` | Displays the built-in help and flag usage reference manual. |

---

## 📦 Configuration Specification (`config.json`)

The application behavior is completely customized using a JSON file. The `report_on` block controls which metrics sections will be active and appended to the final report summary:

```json
{
  "report_path": "./reports/",
  "report_mask": "2006-01-02_15-04",
  "report_on": {
    "report_im": true,
    "report_jmeter": true,
    "report_scenario": true,
    "report_dash": true,
    "report_clickhouse": true
  },
  "allure": {
    "report_allure": true,
    "token": "your-allure-token",
    "report_endpoint": "http://allure-server:8080/api/v1/",
    "report_path": "./allure-results/"
  },
  "jmeter": {
    "jmeter_influx": "http://localhost:8086",
    "jmeter_query": "SELECT mean(value) FROM...",
    "jmeter_query_threshold": [
      {
        "name": "*",
        "error_field": "error_rate",
        "threshold": 1.0,
        "description": "Threshold exceeded! Max allowed: %f, Current: %d in test suite %s"
      }
    ]
  },
  "clickhouse": {
    "server": "ch-server.local:8123",
    "user": "default",
    "pass": "password",
    "query": [
      {
        "dbname": "analytics",
        "sql": "SELECT * FROM lt_metrics WHERE...",
        "name": "ClickHouse Aggregated Analysis"
      }
    ]
  },
  "confluence": {
    "report_confluence_on": true,
    "report_confluence_url": "https://company.com",
    "parent_page_id": "12345678"
  }
}
```

### Grafana Metric Validation Rules:
The engine applies two distinct evaluation strategies for Grafana dashboard targets:
1. **Conditional Expressions (`Comparison`):** Checks whether the source database percentile values violate the `Threshold` setting based on conditional operators (`>` or `<`).
2. **Built-in Dashboard Alerts (`AlertID`):** When given an alert ID, the engine tracks the state history in Grafana and registers a critical incident if an active alert is captured.

---

## 🛠 Compilation and Build Targets

Building the application binary requires **Go 1.18+** and the `make` utility installed on your local workstation.

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

Enabling the `report_confluence_on` option triggers the following publication cycle:
1. Creates a nested child page inside your Confluence directory structure beneath the designated `parent_page_id`.
2. Formats markdown contents layout onto the new page body and attaches the master PDF binary report.
3. Generates a live publication tracking URL, which is automatically synchronized and appended to your Allure test runs.

---

## 👨‍💻 Contribution Guidelines

We highly appreciate community contributions to expand this toolset!
- Spotted a bug or want to submit an improvement proposal? Open an [Issue](https://github.com).
- Ready to fix an issue or submit a feature extension? Open a [Pull Request](https://github.com).

---

## 📄 License

This software is distributed under the terms of the **MIT License**. Check out the complete text in the accompanying [LICENSE](LICENSE) file.

**Author:** [Egor Kovalchuk](https://github.com)  
**Current Version:** `0.6.0.2`