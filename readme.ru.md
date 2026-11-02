# Load Test Report (go-ltreport)
Утилита на Go для автоматизации анализа результатов нагрузочного тестирования и генерации сводных отчетов. 

Инструмент помогает разработчикам, QA-инженерам и DevOps-специалистам быстро превращать сырые метрики из различных источников в структурированные PDF-отчеты или автоматически публиковать их на страницах Confluence и в Allure.

## 🚀 Ключевые возможности

- **Мультиплатформенный сбор метрик:** Интеграция с InfluxDB (JMeter), Grafana, ClickHouse, Prometheus, Graphite, Telegraf и HP Service Manager (HPSM/FSM).
- **Умная аналитика превышений:** Автоматический анализ метрик по заданным пороговым значениям (thresholds) как для тестов целиком, так и по отдельным сценариям/транзакциям. Поддержка операторов сравнения (`>` и `<`).
- **Наглядные PDF-отчеты:** Генерация полноценных документов, содержащих интерактивные графики (скриншоты панелей Grafana), таблицы инцидентов и детальный анализ сценариев нагрузочного тестирования.
- **Шаблонизация Grafana (CSV):** Поддержка массовой параллельной генерации графиков по шаблону на основе входных CSV-файлов. Опциональная настройка `ThresholdFilter` — показывать графики только при срабатывании порога.
- **Автоматизация Confluence:** Автоматическая выгрузка отчетов на целевые страницы Confluence в виде дочерних страниц с прикрепленным PDF-файлом. Поддержка аутентификации по паролю и токену.
- **Интеграция с Allure Framework:** Результаты проверок каждого порога могут экспортироваться в Allure в виде тест-кейсов со статусами `passed`, `failed` или `broken` с прикрепленными артефактами.
- **Поддержка нескольких кодировок:** Автоматическое определение и конвертация Windows-1251, KOI8-R, ISO-8859-5 и UTF-8.
- **Внешние файлы конфигурации:** Поддержка загрузки конфигураций JMeter, Grafana Dashboards и Grafana Templates из отдельных JSON-файлов через поля `JmeterConfig`, `GrafanadashConfig` и `GrafanadashTemplateConfig`.

## 📐 Архитектура проекта и кодовой базы

Кодовая база проекта организована в соответствии с практиками разработки **Standard Go Project Layout**:

* **`cmd/ltreport/`** — Точка входа в приложение. Содержит `go-ltreport.go` (основной обработчик CLI, оркестрация отчетов) и `init.go` (инициализация переменных, обработка времени, управление временными файлами).
* **`internal/`** — Приватный изолированный код приложения.
  * `internal/reportdata/` — Основные структуры данных, загрузка конфигурации, валидация, парсинг порогов и утилиты.
  * `internal/source/` — Модули и адаптеры для подключения к источникам метрик.
    * `internal/source/influx/` — Клиент InfluxDB (JMeter).
    * `internal/source/grafana/` — Клиент Grafana (дашборды и алерты).
    * `internal/source/prometheus/` — Клиент Prometheus.
    * `internal/source/graphite/` — Клиент Graphite API.
    * `internal/source/telegraf/` — Клиент Telegraf.
    * `internal/source/clickhouse/` — Клиент ClickHouse.
  * `internal/pdf/` — Движок генерации отчетов и рендеринга таблиц/графиков в PDF.
  * `internal/allure/` — Клиент для интеграции и отправки результатов в Allure.
  * `internal/confluence/` — Клиент API Confluence для создания страниц и прикрепления PDF.
  * `internal/hpsm/` — Клиент для интеграции с HP Service Manager (FSM).
  * `internal/logger/` — Структурированное логирование с записью в файл и опциональным выводом в консоль.
  * `internal/notify/` — Модуль отправки уведомлений.
  * `internal/fonts/` — Встроенные шрифты (варианты Times New Roman) для рендеринга PDF.
* **`example/`** — Примеры конфигурационных файлов и шаблонов для быстрого старта.

## 🔄 Процесс работы утилиты (Workflow Pipeline)
При запуске приложение выполняет последовательный пайплайн сбора и обработки данных:

```
[Запуск main()] ──> [Парсинг флагов CLI] ──> [Чтение config.json]
│
└──> [Проверка дат и создание tmp/]
│
└──> [Вызов StartReport()]
│
├──> [Инициализация PDF-движка и Allure]
├──> [Параллельный сбор метрик: InfluxDB/JMeter, Grafana, ClickHouse, HPSM]
├──> [Анализ превышения порогов (Thresholds Check) -> Запись ошибок]
├──> [Рендеринг PDF-документа (Таблицы + Скриншоты панелей)]
└──> [Загрузка PDF в Confluence (опционально) + Выгрузка архива в Allure]
│
└──> [Удаление tmp/ (флаг -rm)] ──> [Отправка уведомлений (если не -ban)] ──> [Завершение]
```

---

## ⚙️ Флаги командной строки (CLI Flags)

Полный список параметров, доступных при запуске бинарного файла:

| Флаг | Тип данных | Значение по умолчанию | Описание |
| :--- | :--- | :--- | :--- |
| `-config` | string | `config.json` | Путь к пользовательскому конфигурационному файлу приложения. |
| `-d` | boolean | `false` | Активация режима отладки (`debug mode`) с расширенным логированием. |
| `-v` | boolean | `false` | Вывод текущей версии утилиты (Текущая версия: `0.6.0.2`). |
| `-h` | boolean | `false` | Вывод встроенной справки и списка доступных флагов. |
| `-c` | boolean | `false` | Включает вывод в консоль дополнительно к файлу логирования. |
| `-hour` | boolean | `false` | Генерация автоматического почасового отчета за прошедший час. |
| `-start` | string | `""` | Дата начала отчета в формате `2006.01.02 15:04` (например, `"2026.01.31 15:00"`). |
| `-end` | string | `""` | Дата окончания отчета в формате `2006.01.02 15:04` (например, `"2026.01.31 18:00"`). |
| `-rm` | boolean | `false` | Автоматическое удаление временных PNG-файлов после генерации отчета. |
| `-ban` | boolean | `false` | Запрет на отправку уведомлений об отчете по завершении работы. |
| `-fsmlogin` | string | `""` | Переопределяет логин FSM/HP Service Manager из конфигурационного файла. |
| `-fsmpass` | string | `""` | Переопределяет пароль FSM/HP Service Manager из конфигурационного файла. |
| `-conflproxy` | string | `""` | Устанавливает прокси для Confluence (формат: `http://user:password@host:port`). |
| `-ConfToken` | string | `""` | Переопределяет токен авторизации Confluence из конфигурационного файла. |
| `-CHUser` | string | `""` | Переопределяет имя пользователя ClickHouse из конфигурационного файла. |
| `-CHPassword` | string | `""` | Переопределяет пароль ClickHouse из конфигурационного файла. |

> **Примечание:** Флаги `-fsmlogin`, `-fsmpass`, `-conflproxy`, `-ConfToken`, `-CHUser` и `-CHPassword` имеют приоритет над значениями, указанными в конфигурационном файле.

---

## 📦 Спецификация конфигурации (`config.json`)

Приложение гибко настраивается через JSON-конфиг. Блок `report_on` отвечает за включение или отключение конкретных секций в финальном отчете:

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
    "jmeter_login_influx": "",
    "jmeter_pass_influx": "",
    "jmeter_query": "SELECT mean(\"rate\"), mean(\"avg\"), mean(\"errpct\"), percentile(\"avg\", 95), max(\"avg\"), percentile(\"avg\", 50) FROM \"delta\" WHERE ",
    "jmeter_query_group": "GROUP BY \"suite\", time(1d) fill(0)",
    "jmeter_query_field": [
      {"Name": "Rate", "Description": "Средняя пропускная способность %d ops"},
      {"Name": "Avg", "Description": "Средняя задержка %d ms"},
      {"Name": "ErrorPercent", "Description": "Ошибки %d %%"},
      {"Name": "Percentile", "Description": "Перцентиль задержки %d ms"},
      {"Name": "Maxlatency", "Description": "Максимальная задержка %d ms"},
      {"Name": "Percentile50", "Description": "Перцентиль 50 %d ms"}
    ],
    "jmeter_query_threshold": [
      {"Name": "oapi-get", "ErrorField": "ErrorPercent", "Threshold": 3, "Description": "Превышен уровень ошибок %d в тесте %s текущий %d"}
    ],
    "jmeter_query_scenario": "SELECT max(\"pct90.0\"), max(\"max\"), mean(\"count\") FROM \"details\" WHERE ",
    "jmeter_query_scrn_group": " GROUP BY time(1d), \"application\", \"transaction\", \"statut\" fill(0) ORDER BY time DESC",
    "jmeter_query_scrn_field": [
      {"Name": "Percentile90", "Description": "Перцентиль 90 задержка %d ms"},
      {"Name": "Maxlatency", "Description": "Максимальная задержка %d ms"},
      {"Name": "Rate", "Description": "Средняя пропускная способность %d ops"}
    ],
    "jmeter_query_scrn_threshold": [
      {"Name": "oapi_get", "NameThread": "getToken", "ErrorField": "Percentile90", "Threshold": 5, "Statut": "ok", "Description": "Высокая задержка %d ms в сценарии %s текущий %d ms"}
    ]
  },
  "grafana_dash": [
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
      "ThDescription": "Превышен порог CPU",
      "Tag": "Product:Test;layer:system",
      "Size": {"Width": 120, "Height": 50},
      "ThresholdFilter": false,
      "FileList": "",
      "FilePettern": ""
    }
  ],
  "grafanadash_template": [
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
      "ThDescription": "Превышен порог CPU",
      "Tag": "Product:Test",
      "Size": {"Width": 120, "Height": 50},
      "ThresholdFilter": true,
      "FileList": "templates.csv",
      "FilePettern": "server;name"
    }
  ],
  "clickhouse": {
    "server": "ch-server.local:8123",
    "user": "default",
    "pass": "password",
    "query": [
      {
        "dbname": "analytics",
        "sql": "SELECT * FROM lt_metrics WHERE...",
        "name": "ClickHouse Сводный Анализ"
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

### Валидация метрик Grafana в коде:
Движок применяет две различные стратегии оценки для целей дашборда Grafana:
1. **Условные выражения (`Comparison`):** Проверяет, нарушают ли значения перцентиля из исходной базы данных настройку `Threshold`, основываясь на условных операторах (`>` или `<`). Поле `Threshold` может быть либо числовым значением (по умолчанию используется сравнение `>`), либо строкой вроде `"<10"` или `">90"` для явного задания оператора сравнения.
2. **Встроенные алерты дашборда (`AlertID`):** При указании ID алерта движок отслеживает историю состояний в Grafana и регистрирует критический инцидент, если обнаружен активный алерт.

### Справочник типов источников для дашбордов Grafana:
| SourceType | Источник данных |
|:---:|---|
| `1` | Telegraf / InfluxDB |
| `2` | Prometheus |
| `3` | Graphite |

---

> **Примечание:** Конфигурация поддерживает внешние файлы для JMeter, Grafana Dashboards и Grafana Templates через поля `JmeterConfig`, `GrafanadashConfig` и `GrafanadashTemplateConfig` соответственно. Это позволяет поддерживать конфигурации модульными.

---

## 🛠 Сборка проекта

Для компиляции проекта вам понадобятся установленный **Go 1.24+** и утилита `make`.

* **Сборка бинарного файла:**
  ```bash
  make build
  ```
* **Запуск тестов:**
  ```bash
  make test
  ```
* **Очистка временных файлов сборки:**
  ```bash
  make clean
  ```

---

## 📖 Примеры использования

### 1. Стандартный отчет за день (по умолчанию берется `config.json`)
```bash
./ltreport
```

### 2. Отчет за произвольный период времени с кастомным конфигом
```bash
./ltreport -start "2026.12.31 09:00" -end "2026.12.31 18:00" -config custom_config.json
```

### 3. Запуск в режиме отладки с очисткой папки tmp по завершении
```bash
./ltreport -d -rm -config config.json
```

---

## 🤝 Интеграция с Confluence

Включение опции `ReportConfluenceOn` запускает следующий цикл публикации:
1. Создаёт вложенную дочернюю страницу в структуре вашего пространства Confluence под назначенным `ReportConfluenceId`.
2. Форматирует содержимое markdown на теле новой страницы и прикрепляет мастер-файл PDF-отчета.
3. Генерирует live-ссылку на публикацию, которая автоматически синхронизируется и добавляется к вашим тестовым запускам в Allure.

Аутентификация поддерживает как **метод на основе пароля** (`ReportConfluenceLogin` / `ReportConfluencePass`), так и **метод на основе токена** (`ReportConfluenceToken`). Прокси можно указать через `ReportConfluenceProxy`.

---

## 👨‍💻 Разработка и Contribution

Мы рады вашей помощи в развитии проекта!
- Нашли баг или есть идея? Создайте [Issue](https://github.com).
- Хотите улучшить код? Отправляйте [Pull Request](https://github.com).

---

## 📄 Лицензия

Проект распространяется на условиях лицензии **MIT**. Подробнее см. в файле [LICENSE](LICENSE).

**Автор:** [Egor Kovalchuk](https://github.com)  
**Текущая версия:** `0.6.0.2`