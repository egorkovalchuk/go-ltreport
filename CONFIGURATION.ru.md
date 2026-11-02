# Конфигурационный файл LTReport

## Основные настройки отчета
```json
{
    "ReportFilename": "LTReport",
    "ReportMask": "02012006", 
    "ReportPath": "/tmp"
}
```

**Описание параметров:**
- `ReportFilename` - базовое имя генерируемого отчета
- `ReportMask` - маска для форматирования даты в имени файла (DDMMYYYY)
- `ReportPath` - путь для сохранения отчетов

---

## Настройки Confluence
```json
"Confluence": {
    "ReportConfluenceOn": true,
    "ReportConfluenceURL": "https://confluence.com/",
    "ReportConfluenceId": "926291954",
    "ReportConfluenceSpace": "space",
    "ReportConfluenceLogin": "login",
    "ReportConfluencePass": "pass",
    "ReportConfluenceToken": "token",
    "ReportConfluenceProxy": ""
}
```

**Описание параметров:**
- `ReportConfluenceOn` - включить загрузку отчетов в Confluence
- `ReportConfluenceURL` - URL сервера Confluence
- `ReportConfluenceId` - ID целевой страницы
- `ReportConfluenceSpace` - пространство Confluence  
- `ReportConfluenceLogin/Pass/Token` - учетные данные для авторизации
- `ReportConfluenceProxy` - прокси для подключения к Confluence

---

## Настройки Allure
```json
"Allure": {
    "ReportAllure": true,
    "ReportEndPoint": "https://allure.company.com",
    "Token": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "ReportPath": "allure-results/",
    "ReportAllureProxy": ""
}
```

**Описание параметров:**
- `ReportAllure` - включить интеграцию с Allure
- `ReportEndPoint` - URL сервера Allure
- `Token` - токен авторизации Allure
- `ReportPath` - путь к результатам Allure
- `ReportAllureProxy` - прокси для подключения к Allure

---

## Настройки FSM (Service Manager)
```json
"ReportIM": {
    "Token": "",
    "LoginFSM": "string",
    "PassFSM": "string", 
    "FSMConnect": "fsm.ru",
    "FSMTU": "string"
}
```

**Описание параметров:**
- `LoginFSM/PassFSM` - учетные данные для FSM
- `FSMConnect` - адрес сервера FSM
- `FSMTU` - идентификатор технического устройства
- `Token` - токен авторизации FSM

---

## Настройки ClickHouse
```json
"ClickHouse": {
    "Server": "localhost:1234",
    "User": "stress_test", 
    "Pass": "",
    "Token": "",
    "Query": [
        {
            "sql": "",
            "dbname": "",
            "name": ""
        }
    ]
}
```

**Описание параметров:**
- `Server` - адрес сервера ClickHouse (host:port)
- `User/Pass/Token` - учетные данные для подключения
- `Query` - массив SQL запросов для выполнения:
  - `sql` - SQL запрос
  - `dbname` - имя базы данных
  - `name` - название запроса

---

## Включение модулей отчетности
```json
"ReportOn": {
    "ReportJmeter": true,
    "ReportIM": true, 
    "ReportScenario": true,
    "ReportDash": true,
    "ClickHouse": true
}
```

**Описание параметров:**
- `ReportJmeter` - включить отчет из JMeter
- `ReportIM` - включить отчет из FSM
- `ReportScenario` - включить отчет по сценариям
- `ReportDash` - включить отчет по дашбордам
- `ClickHouse` - включить отчет из ClickHouse

---

## Внешние файлы конфигурации

Конфигурация поддерживает вынос JMeter, Grafana Dashboards и Grafana Templates в отдельные файлы через поля `JmeterConfig`, `GrafanadashConfig` и `GrafanadashTemplateConfig`. Это позволяет поддерживать конфигурации модульными и переиспользуемыми.

---

## Настройки JMeter
```json
"Jmeter": {
    "JmeterInflux": "http://10.10.10.1:8086",
    "JmeterLoginInflux": "",
    "JmeterPassInflux": "",
    "JmeterQuery": "SELECT mean(\"rate\"), mean(\"avg\"), mean(\"errpct\"), percentile(\"avg\", 95), max(\"avg\"), percentile(\"avg\", 50) FROM \"delta\" WHERE ",
    "JmeterQueryGroup": "GROUP BY \"suite\", time(1d) fill(0)",
    "JmeterQueryField": [
        {"Name": "Rate", "Description": "Средняя пропускная способность %d ops"},
        {"Name": "Avg", "Description": "Средняя задержка %d ms"},
        {"Name": "ErrorPercent", "Description": "Ошибки %d %%"},
        {"Name": "Percentile", "Description": "Перцентиль задержки %d ms"},
        {"Name": "Maxlatency", "Description": "Максимальная задержка %d ms"},
        {"Name": "Percentile50", "Description": "Перцентиль 50 %d ms"}
    ],
    "JmeterQueryThreshold": [
        {"Name": "oapi-get", "ErrorField": "ErrorPercent", "Threshold": 3, "Description": "Превышен уровень ошибок %d в тесте %s текущий %d"}
    ],
    "JmeterQueryScenario": "SELECT max(\"pct90.0\"), max(\"max\"), mean(\"count\") FROM \"details\" WHERE ",
    "JmeterQueryScnrGroup": " GROUP BY time(1d), \"application\", \"transaction\", \"statut\" fill(0) ORDER BY time DESC",
    "JmeterQueryScnrField": [
        {"Name": "Percentile90", "Description": "Перцентиль 90 задержка %d ms"},
        {"Name": "Maxlatency", "Description": "Максимальная задержка %d ms"},
        {"Name": "Rate", "Description": "Средняя пропускная способность %d ops"}
    ],
    "JmeterQueryScnrThreshold": [
        {"Name": "oapi_get", "NameThread": "getToken", "ErrorField": "Percentile90", "Threshold": 5, "Statut": "ok", "Description": "Высокая задержка %d ms в сценарии %s текущий %d ms"}
    ]
}
```

**Описание параметров JMeter:**
- `JmeterInflux` - URL InfluxDB с метриками JMeter
- `JmeterLoginInflux/JmeterPassInflux` - учетные данные InfluxDB
- `JmeterQuery` - основной запрос для получения метрик
- `JmeterQueryGroup` - группировка для основного запроса
- `JmeterQueryField` - описание полей результатов запроса (массив объектов Name + Description)
- `JmeterQueryThreshold` - пороговые значения для проверки метрик
- `JmeterQueryScenario` - запрос для анализа сценариев
- `JmeterQueryScnrGroup` - группировка для запроса сценариев
- `JmeterQueryScnrField` - описание полей сценариев (массив объектов Name + Description)
- `JmeterQueryScnrThreshold` - пороговые значения для сценариев (поддерживает NameThread, Statut для фильтрации)

---

## Настройки Grafana Dashboards
```json
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
        "ThDescription": "Превышен порог CPU",
        "Tag": "Product:Test;layer:system",
        "Size": {"Width": 120, "Height": 50},
        "ThresholdFilter": false,
        "FileList": "",
        "FilePettern": ""
    }
]
```

**Описание параметров Grafana:**
- `Name` - название дашборда
- `UrlDash` - URL дашборда Grafana
- `AuthHeader` - заголовок авторизации Bearer token
- `UrlImg` - URL для рендеринга изображения
- `Query` - запрос для получения данных
- `SourceType` - тип источника данных (1-telegraf, 2-prometheus, 3-graphite)
- `UrlQuery` - URL для выполнения запросов к данным
- `UrlQueryGroup` - группировка для запросов
- `Threshold` - пороговое значение для триггера, может принимать на вход условие сравнения, например "<10" - меньше 10, или ">90" - больше 90
- `ThDescription` - описание порога
- `Tag` - тег для категоризации, формат `key:value;key:value` (передаётся в Allure как Label)
- `Size` - размер изображения в отчете (ширина x высота)
- `ThresholdFilter` - показывать график только при срабатывании порога
- `FileList` - файл со списком элементов для массовой генерации
- `FilePettern` - шаблон для именовании элементов в FileList

---

## Шаблоны Grafana Dashboards
```json
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
        "ThDescription": "Превышен порог CPU",
        "Tag": "Product:Test",
        "Size": {"Width": 120, "Height": 50},
        "ThresholdFilter": true,
        "FileList": "templates.csv",
        "FilePettern": "server;name"
    }
]
```

**Дополнительные параметры шаблонов:**
- `Tag` - Описывает данные Label для Allure в формате key:value;key:value...
- `ThresholdFilter` - показывать график только при превышении порога
- `FileList` - файл со списком элементов для шаблона (CSV с разделителем по умолчанию `;`)
- `FilePettern` - шаблон для именовании файлов, формат `column1;column2` (имена столбцов в CSV, используемых для генерации)
