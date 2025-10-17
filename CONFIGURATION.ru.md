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
    "ReportEndPoint": "https://allure.ru",
    "Token": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "ReportPath": "allure-results/"
}
```

**Описание параметров:**
- `ReportAllure` - включить интеграцию с Allure
- `ReportEndPoint` - URL сервера Allure
- `Token` - токен авторизации Allure
- `ReportPath` - путь к результатам Allure

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
    "ClickHouse": false
}
```

**Описание параметров:**
- `ReportJmeter` - включить отчет из JMeter
- `ReportIM` - включить отчет из FSM
- `ReportScenario` - включить отчет по сценариям
- `ReportDash` - включить отчет по дашбордам
- `ClickHouse` - включить отчет из ClickHouse

---

## Настройки JMeter
```json
"Jmeter": {
    "JmeterInflux": "http://10.10.10.1:8086",
    "LoginInflux": "",
    "PassInflux": "",
    "JmeterQuery": "SELECT mean(\"rate\"), mean(\"avg\"), mean(\"errpct\"), percentile(\"avg\", 95), max(\"avg\"), percentile(\"avg\", 50) FROM \"delta\" WHERE ",
    "JmeterQueryGroup": "GROUP BY \"suite\", time(1d) fill(0)",
    
    "JmeterQueryField": [
        {
            "Name": "Rate",
            "Description": "Averagate rate %d ops"
        }
    ],
    
    "JmeterQueryThreshold": [
        {
            "Name": "oapi-get",
            "ErrorField": "ErrorPercent", 
            "Threshold": 3,
            "Description": "The error is higher %d in test %s current %d"
        }
    ],
    
    "JmeterQueryScenario": "SELECT max(\"pct90.0\"), max(\"max\"), mean(\"count\") FROM \"details\" WHERE ",
    "JmeterQueryScnrGroup": " GROUP BY time(1d), \"application\", \"transaction\", \"statut\" fill(0) ORDER BY time DESC",
    
    "JmeterQueryScnrField": [
        {
            "Name": "Percentile90",
            "Description": "Percentile90 latency %d ms"
        }
    ],
    
    "JmeterQueryScnrThreshold": [
        {
            "Name": "oapi_get",
            "NameThread": "getToken",
            "ErrorField": "Percentile90",
            "Threshold": 5,
            "Statut": "ok",
            "Description": "The high latency %d ms in scenario %s current %d ms"
        }
    ]
}
```

**Описание параметров JMeter:**
- `JmeterInflux` - URL InfluxDB с метриками JMeter
- `LoginInflux/PassInflux` - учетные данные InfluxDB
- `JmeterQuery` - основной запрос для получения метрик
- `JmeterQueryGroup` - группировка для основного запроса
- `JmeterQueryField` - описание полей результатов запроса
- `JmeterQueryThreshold` - пороговые значения для проверки метрик
- `JmeterQueryScenario` - запрос для анализа сценариев
- `JmeterQueryScnrGroup` - группировка для запроса сценариев
- `JmeterQueryScnrField` - описание полей сценариев
- `JmeterQueryScnrThreshold` - пороговые значения для сценариев

---

## Настройки Grafana Dashboards
```json
"GrafanaDash": [
    {
        "Name": "ESB CPU",
        "UrlDash": "https://grafana.ru/d/000000879/wf?orgId=1&refresh=5m",
        "AuthHeader": "Bearer Hash=",
        "UrlPanel": "",
        "UrlImg": "https://grafana.ru/render/d-solo/000000879/wf?height=500&orgId=1&panelId=88&refresh=5m&tz=Europe%2FMoscow&width=1000",
        "Query": "SELECT percentile(\"usage_active\",95) FROM \"ret_30w\".\"cpu\" WHERE (\"host\" =~ /server_name[0-2]{1}[0-9]{1}/ AND \"cpu\" = 'cpu-total') ",
        "SourceType": 1,
        "UrlQuery": "https://grafana.ru/api/datasources/proxy/459/query?db=telegraf&q=",
        "UrlQueryGroup": "GROUP BY time(1d) fill(none)",
        "Threshold": 90,
        "ThDescription": "CPU jopa",
        "Tag": "Product",
        "Size": {
            "Width": 120,
            "Height": 50
        }
    }
]
```

**Описание параметров Grafana:**
- `Name` - название дашборда
- `UrlDash` - URL дашборда Grafana
- `AuthHeader` - заголовок авторизации Bearer token
- `UrlPanel` - URL конкретной панели
- `UrlImg` - URL для рендеринга изображения
- `Query` - запрос для получения данных
- `SourceType` - тип источника данных (1-telegraf, 2-prometheus, 3-graphite)
- `UrlQuery` - URL для выполнения запросов к данным
- `UrlQueryGroup` - группировка для запросов
- `Threshold` - пороговое значение для триггера
- `ThDescription` - описание порога
- `Tag` - тег для категоризации
- `Size` - размер изображения в отчете (ширина x высота)

---

## Шаблоны Grafana Dashboards
```json
"GrafanadashTemplate": [
    {
        "Name": "ESB CPU",
        "UrlDash": "https://grafana.ru/d/000000879/wf?orgId=1&refresh=5m",
        "AuthHeader": "Bearer Hash=",
        "UrlPanel": "",
        "UrlImg": "https://grafana.ru/render/d-solo/000000879/wf?height=500&orgId=1&panelId=88&refresh=5m&tz=Europe%2FMoscow&width=1000",
        "Query": "SELECT percentile(\"usage_active\",95) FROM \"ret_30w\".\"cpu\" WHERE (\"host\" =~ /server_name[0-2]{1}[0-9]{1}/ AND \"cpu\" = 'cpu-total') ",
        "SourceType": 1,
        "UrlQuery": "https://grafana.ru/api/datasources/proxy/459/query?db=telegraf&q=",
        "UrlQueryGroup": "GROUP BY time(1d) fill(none)",
        "Threshold": 90,
        "ThDescription": "CPU jopa",
        "Size": {
            "Width": 120,
            "Height": 50
        },
        "ThresholdFilter": true,
        "Tag": "Product",
        "FileList": "1.csv",
        "FilePettern": "server;name"
    }
]
```

**Дополнительные параметры шаблонов:**
- `Tag` - Описывает данные Label для Allure в формате key:value;ke:value...
- `ThresholdFilter` - показывать график только при превышении порога
- `FileList` - файл со списком элементов для шаблона
- `FilePettern` - шаблон для именования файлов