# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0.3] - 2026-11-02

### Added
- Unit tests for all major packages (allure, confluence, hpsm, logger, notify, reportdata, clickhouse, grafana, graphite, influx, prometheus, telegraf)
- Example configurations (grafana-dash, grafana-template, jmeter, servers)

### Changed
- Updated README and CONFIGURATION documentation

### Fixed
- Allure token retrieval

## [0.6.0.2] - 2024-10-29

### Added
- Makefile

## [0.6.0.1] - 2024-10-29

### Added
- Comparison type parameter for threshold configuration
- Calculation of the 99th percentile when a vector of values is received
- Mandatory range for Prometheus queries (`[range]`)

### Fixed
- Fixed threshold type to `float64`

## [0.6.0.0] - 2024-10-29

### Added
- Alert data retrieval from Grafana
- Attachment support for Allure reports

## [0.5.0.1] - 2024-10-29

### Added
- UTF-8 support
- Font embedding in PDF reports

### Fixed
- String conversion to UTF-8

## [0.5.0.0] - 2024-10-29

### Added
- Allure report generation
- Goroutine support
- HPSM API integration (local custom)

### Fixed
- Directory validation

## [0.4.0.1] - 2024-10-29

### Added
- Notifier function

## [0.4.0.0] - 2024-10-29

### Added
- Grafana template
- Graphite API support

## [0.3.0.2] - 2024-10-29

### Changed
- Removed PNG files generation

## [0.3.0.1] - 2024-10-29

### Added
- PDF report support for ClickHouse

## [0.3.0.0] - 2024-10-29

### Added
- ClickHouse support

## [0.2.0.1] - 2024-10-29

### Added
- Token authentication for Confluence

## [0.2.0.0] - 2024-10-29

### Added
- Custom time period support
