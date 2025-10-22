# **go-ltreport** 📊  
**Генератор отчётов для нагрузочного тестирования**  

Утилита для автоматической генерации отчётов на основе данных из:  
- **JMeter** (через InfluxDB)  
- **Grafana** (дашборды и графики)  
- **ClickHouse** (аналитика)  
- **Confluence** (публикация результатов)  

---

## **🚀 Возможности**  
- 📌 Автоматическая сборка отчётов в PDF  
- 📊 Визуализация метрик из Grafana  
- ⚡ Проверка SLA (пороговые значения)  
- 📌 Публикация в Confluence  
- 🔍 Поддержка нескольких источников данных (InfluxDB, ClickHouse)  

---

## **🛠 Установка**  
### **1. Сборка из исходников**  
```bash 
git clone https://github.com/egorkovalchuk/go-ltreport.git
cd go-ltreport
go build -o ltreport main.go init.go pdf.go 
```

### **2. Запуск**  
```bash 
./ltreport -config config.json
```

---

## **⚙️ Конфигурация**  
Пример `config.json`:  
```json
{
  "ReportFilename": "LTReport",
  "ReportConfluenceOn": true,
  "ReportConfluenceURL": "https://confluence.example.com",
  "JmeterInflux": "http://localhost:8086",
  "GrafanaDash": [
    {
      "Name": "CPU Usage",
      "UrlDash": "https://grafana.example.com/d/abcd1234",
      "Threshold": 90
    }
  ]
}
```

📌 **Обязательные параметры:**  
- `ReportFilename` — имя отчёта.  
- `JmeterInflux` — URL InfluxDB с метриками JMeter.  
- `GrafanaDash` — список дашбордов Grafana для выгрузки.  

---

## **📄 Примеры**  
### **1. Генерация PDF-отчёта**  
``` bash
# Стандартный отчет за день
./ltreport -c config.json

# Отчет за произвольный период
./ltreport -start "2023.12.31 09:00" -end "2023.12.31 18:00"

# Режим отладки
./ltreport -d -c custom_config.json
``` 

### **2. Публикация в Confluence**  
```json
{
  "ReportConfluenceOn": true,
  "ReportConfluenceId": "12345",
  "ReportConfluenceSpace": "LOADTESTS"
}
```  
---

## **📌 Логирование**  
Логи сохраняются в `ltreport.log`:  
```text
INFO: Config loaded successfully  
ERROR: InfluxDB connection failed (timeout)  
```  

---

## **📌 Лицензия**  
MIT License.  

---

## **🤝 Как помочь проекту?**  
1. 🐞 Сообщайте о багах в **Issues**.  
2. 💡 Предлагайте улучшения через **Pull Requests**.  
3. 📢 Расскажите о проекте!  

---

**📌 Ссылки:**  
- [Пример конфига](config.example)  
- [Документация](docs/) (в разработке)  

--- 

**Автор:** [Egor Kovalchuk](https://github.com/egorkovalchuk)  
**Версия:** 0.6.0