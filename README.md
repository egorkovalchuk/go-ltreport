# Load Test Report
Utility for Generating Load Testing Reports

## Project Description
LT report is a tool designed to automate the analysis and generation of reports based on load testing results. This project is aimed at developers, QA engineers, and DevOps specialists who need to quickly convert raw metrics into structured reports in a convenient format.

A utility for automating the generation of comprehensive reports based on load testing results, with integration into:
* InfluxDB/JMeter
* Grafana
* ClickHouse
* Prometheus
* Confluence
* HP Service Manager

## Key Features
* Automatic collection of metrics from various sources
* Analysis of threshold value exceedances
* Generation of PDF reports with charts and tables
* Uploading reports to Confluence
* Support for custom time periods

## Usage
* Use -v gor get version
* Use -d start with debug mode
* Use -c start with users config
* Use -hour to generate an hourly report
* Use -fsmlogin start with Login FSM(not suppoted)
* Use -fsmpass start with Password FSM(not suppoted)
* Use -conflproxy start with proxy for connection to Confluence
* Use -start start custom date (format: 2006.01.02 15:04 )
* Use -end end custom date (format: 2006.01.02 15:04 )
* Use -ConfToken Token for Confluence access
* Use -CHUser ClickHouse username
* Use -CHPass ClickHouse password

## Report Generation
### Structure of the PDF Report
* Title Page with Date
* Summary of Incidents
* Grafana Charts
* Data from ClickHouse
* Detailed Test Statistics
* Analysis of Load Testing Scenarios

## Integration with Confluence
To upload reports, you need to:
* Specify credentials in the configuration file.
* Set the ID of the target page.
* Ensure write permissions are granted.
Reports are automatically created as child pages in Confluence with attached PDF files.

## Example 
``` bash
# Стандартный отчет за день
./ltreport -c config.json

# Отчет за произвольный период
./ltreport -start "2023.12.31 09:00" -end "2023.12.31 18:00"

# Режим отладки
./ltreport -d -c custom_config.json
```
