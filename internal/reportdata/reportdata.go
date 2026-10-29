package reportdata

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
)

// Config configuration stucture
type Config struct {
	// Имя файла
	ReportFilename string `json:"ReportFilename"`
	// Маска даты
	ReportMask string `json:"ReportMask"`
	// Путь к отчету
	ReportPath string         `json:"ReportPath"`
	Confluence ConfluenceType `json:"Confluence"`
	Allure     struct {
		ReportAllure      bool   `json:"ReportAllure"`
		ReportEndPoint    string `json:"ReportEndPoint"`
		Token             string `json:"Token"`
		ReportPath        string `json:"ReportPath"`
		ReportAllureProxy string `json:"ReportAllureProxy,omitempty"`
	} `json:"Allure"`
	ReportIM struct {
		Token      string `json:"Token"`
		LoginFSM   string `json:"LoginFSM"`
		PassFSM    string `json:"PassFSM"`
		Fsmconnect string `json:"FSMConnect"`
		Fsmtu      string `json:"FSMTU"`
	} `json:"ReportIM"`
	ClickHouse struct {
		Server string `json:"Server"`
		User   string `json:"User"`
		Pass   string `json:"Pass"`
		Token  string `json:"Token"`
		Query  []struct {
			Sql    string `json:"sql"`
			DBname string `json:"dbname"`
			Name   string `json:"name"`
		} `json:"Query"`
	} `json:"ClickHouse"`
	ReportOn struct {
		//Включить отчет из Jmeter
		ReportJmeter bool `json:"ReportJmeter"`
		//Включить отчет из FSM
		ReportIM bool `json:"ReportIM"`
		//Включить отчет по Сценариям
		ReportScenario bool `json:"ReportScenario"`
		//Включить отчет по дашбордам
		ReportDash       bool `json:"ReportDash"`
		ReportClickHouse bool `json:"ClickHouse"`
	} `json:"ReportOn"`
	JmeterConfig              string   `json:"JmeterConfig"`
	GrafanadashConfig         []string `json:"GrafanadashConfig"`
	GrafanadashTemplateConfig []string `json:"GrafanadashTemplateConfig"`
	// Сохраняем поддержку старых версий
	Jmeter              JMeterType                  `json:"Jmeter"`
	Grafanadash         []GrafanaDashStruct         `json:"GrafanaDash"`
	GrafanadashTemplate []GrafanadashTemplateStruct `json:"GrafanadashTemplate"`
}

type JMeterType struct {
	JmeterLoginInflux string `json:"JmeterLoginInflux"`
	JmeterPassinflux  string `json:"JmeterPassInflux"`
	//Подключение к инфлюксу jmeter
	JmeterInflux string `json:"JmeterInflux"`
	//запрос
	JmeterQuery string `json:"JmeterQuery"`
	//группировка
	JmeterQueryGroup string `json:"JmeterQueryGroup"`
	//Описание полей
	JmeterQueryField []struct {
		//имя поля
		Name string `json:"Name"`
		//Описание
		Description string `json:"Description"`
	} `json:"JmeterQueryField"`
	//Тестовые сценари с порогами по ошибкам
	JmeterQueryThreshold []struct {
		//имя сценария
		Name string `json:"Name"`
		//Поле по которому смотрим пороги
		ErrorField string `json:"ErrorField"`
		//порог
		Threshold float64 `json:"Threshold"`
		//Описание порога
		Description string `json:"Description"`
	} `json:"JmeterQueryThreshold"`
	JmeterQueryScenario      string              `json:"JmeterQueryScenario"`
	JmeterQueryScnrGroup     string              `json:"JmeterQueryScnrGroup"`
	JmeterQueryScnrField     []JmeterQScnrFieldS `json:"JmeterQueryScnrField"`
	JmeterQueryScnrThreshold []struct {
		//имя сценария
		Name string `json:"Name"`
		//имя нити
		NameThread string `json:"NameThread"`
		//Поле по которому смотрим порогиN
		ErrorField string `json:"ErrorField"`
		//Статус на Jmeter
		Statut string `json:"Statut"`
		//порог
		Threshold float64 `json:"Threshold"`
		//Описание порога
		Description string `json:"Description"`
	} `json:"JmeterQueryScnrThreshold"`
}

type ConfluenceType struct {
	// Включение выкладнки на конфлюенс
	ReportConfluenceOn bool `json:"ReportConfluenceOn"`
	// Адрес конфлюенса
	ReportConfluenceURL string `json:"ReportConfluenceURL"`
	// Ид куда пишем данные
	ReportConfluenceId string `json:"ReportConfluenceId"`
	// Спейс конфлюенса
	ReportConfluenceSpace string `json:"ReportConfluenceSpace"`
	ReportConfluenceLogin string `json:"ReportConfluenceLogin"`
	ReportConfluencePass  string `json:"ReportConfluencePass"`
	ReportConfluenceToken string `json:"ReportConfluenceToken"`
	ReportConfluenceProxy string `json:"ReportConfluenceProxy,omitempty"`
}

type GrafanaDashStruct struct {
	Name string `json:"Name"`
	// авторизация на графане
	AuthHeader string `json:"AuthHeader"`
	//ссылка на даш
	Urldash string `json:"UrlDash"`
	//ссылка на панель
	Urlpanel string `json:"UrlPanel"`
	//Ссылка на картинку в графане. Время from to не указвать, она формируется в скрипте
	Urlimg string `json:"UrlImg"`
	//Источник данных
	SourceType int `json:"SourceType"`
	//запрос данных для даша
	Query string `json:"Query"`
	//Порог для запроса
	Threshold     float64
	ThresholdDesc interface{} `json:"Threshold"`
	Comparison    string
	//Описание порога
	ThDescription string `json:"ThDescription"`
	//ссылка на запрос данных, смотреть в графане
	UrlQuery string `json:"UrlQuery"`
	// Tags
	Tag string `json:"Tag"`
	// AlertID
	AlertID any `json:"AlertID"`
	//группировка
	UrlQueryGroup string `json:"UrlQueryGroup"`
	Size          struct {
		Width  int
		Height int
	}
}

type GrafanadashTemplateStruct struct {
	Name string `json:"Name"`
	// авторизация на графане
	AuthHeader string `json:"AuthHeader"`
	// ссылка на даш
	Urldash string `json:"UrlDash"`
	// ссылка на панель
	Urlpanel string `json:"UrlPanel"`
	// Ссылка на картинку в графане. Время from to не указвать, она формируется в скрипте
	Urlimg string `json:"UrlImg"`
	// Источник данных
	SourceType int `json:"SourceType"`
	// запрос данных для даша
	Query string `json:"Query"`
	// Порог для запроса
	Threshold float64 `json:"Threshold"`
	// Описание порога
	ThDescription string `json:"ThDescription"`
	// ссылка на запрос данных, смотреть в графане
	UrlQuery string `json:"UrlQuery"`
	// Tags
	Tag string `json:"Tag"`
	// группировка
	UrlQueryGroup string `json:"UrlQueryGroup"`
	Size          struct {
		Width  int
		Height int
	}
	// Добавлять график если порог стреляет
	ThresholdFilter bool `json:"ThresholdFilter"`
	// Список
	FileList    string `json:"FileList"`
	FilePettern string `json:"FilePettern"`
}

// Вынесена структар из конфига
type JmeterQScnrFieldS struct {
	//имя поля
	Name string `json:"Name"`
	//Описание
	Description string `json:"Description"`
}

// сруктура ошибок для анализа
type LTError struct {
	Name        string
	Threshold   float64
	Description string
	Type        string
	Tag         string
}
type LTErrors []LTError

// сруктура для вывода графиков
type LTGrag struct {
	Name        string
	Threshold   float64
	Description string
	ContentType string
	UrlDash     string
	Size        struct {
		Width  int
		Height int
	}
}
type LTGrags []LTGrag

// Формирование списка для динамических шаблонов
type DinamicRecord map[string]string

func (lt *LTErrors) AddError(name string, threshold float64, description string, percentile float64, tp, tag string) {
	ltp := LTError{Name: "Grafana: " + name, Threshold: threshold, Description: description + ":\n Threshold " + fmt.Sprint(threshold) + " - current " + fmt.Sprint(percentile), Type: tp, Tag: tag}
	*lt = append(*lt, ltp)
}

func Helpstart() {
	fmt.Println("Use -v get version")
	fmt.Println("Use -d start with debug mode")
	fmt.Println("Use -config start with users config")
	fmt.Println("Use -hour to generate an hourly report ")
	fmt.Println("Use -fsmlogin start with Login FSM")
	fmt.Println("Use -fsmpass start with Password FSM")
	fmt.Println("Use -conflproxy start with proxy for connection to Confluence, example http://user:password@url:port")
	fmt.Println("Use -start and -end for generate a report on an arbitrary date ")
}

// Чтение конфига
func (cfg *Config) Readconf(confname string) error {

	file, err := openFile(confname)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &cfg)

	if err != nil {
		return errorJSON(err, file, confname)
	}

	if cfg.JmeterConfig != "" {
		file, err = openFile(cfg.JmeterConfig)
		if err != nil {
			return err
		}
		err = json.Unmarshal(file, &cfg.Jmeter)
		if err != nil {
			return errorJSON(err, file, cfg.JmeterConfig)
		}
	}

	if isEmptyStruct(cfg.Jmeter) {
		fmt.Println("Jmeter bloks is empty")
	}

	if len(cfg.GrafanadashConfig) != 0 {
		for _, i := range cfg.GrafanadashConfig {
			file, err = openFile(i)
			if err != nil {
				return err
			}
			var tmp []GrafanaDashStruct
			err = json.Unmarshal(file, &tmp)
			if err != nil {
				return errorJSON(err, file, i)
			}

			cfg.Grafanadash = append(cfg.Grafanadash, tmp...)
		}
	}

	if len(cfg.GrafanadashTemplateConfig) != 0 {
		for _, i := range cfg.GrafanadashTemplateConfig {
			file, err = openFile(i)
			if err != nil {
				return err
			}
			var tmp []GrafanadashTemplateStruct
			err = json.Unmarshal(file, &tmp)
			if err != nil {
				return errorJSON(err, file, i)
			}

			cfg.GrafanadashTemplate = append(cfg.GrafanadashTemplate, tmp...)
		}
	}

	for i, j := range cfg.Grafanadash {
		th, com, err := cfg.parseThreshold(j.ThresholdDesc)
		if err != nil {
			fmt.Println(err)
			return err
		}
		cfg.Grafanadash[i].Threshold = th
		cfg.Grafanadash[i].Comparison = com
	}

	for _, t := range cfg.GrafanadashTemplate {
		if _, err := os.Stat(t.FileList); os.IsNotExist(err) {
			fmt.Printf("Error:  file '%s' does not exist\n", t.FileList)
		}

	}
	return nil
}

func (cfg *Config) parseThreshold(in interface{}) (th float64, com string, err error) {

	switch v := in.(type) {
	case string:

		Pattern := `([<>]?)(\d+)`

		re, err := regexp.Compile(Pattern)
		if err != nil {
			return 0, "", fmt.Errorf("invalid regex pattern: %w", err)
		}

		matches := re.FindStringSubmatch(v)
		if matches == nil {
			return 0, "", fmt.Errorf("invalid  format: %s", v)
		}

		if len(matches) > 1 {
			if matches[1] != "" {
				com = matches[1]
			} else {
				com = ">"
			}
		}
		if len(matches) > 2 && matches[2] != "" {
			tmp, err := strconv.ParseFloat(matches[2], 64)
			th = tmp
			if err != nil {
				return 0, "", fmt.Errorf("invalid convert threshold: %s", v)
			}
		}
	case int:
		th = float64(v)
		com = ">"
	case float64:
		th = v
		com = ">"
	default:
		th = 0
		com = ""
	}

	return th, com, nil
}

// offsetToLineCol вычисляет строку и колонку по смещению в байтах
func offsetToLineCol(data []byte, offset int64) (line, col int) {
	if offset < 0 {
		return 1, 1
	}
	line = 1
	col = 1 // колонки нумеруются с 1
	// Проходим по каждому байту до указанного смещения (не включая offset)
	for i := 0; i < int(offset) && i < len(data); i++ {
		if data[i] == '\n' {
			line++
			col = 1 // после перевода строки новая колонка
		} else {
			col++
		}
	}
	return line, col + 1 // +1 т.к. колонки начинаются с 1
}

func openFile(filename string) ([]byte, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Error: config file '%s' does not exist\n", filename)
	}

	// Проверяем существование файла
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Файл %s не существует\n", filename)
		}
		return nil, fmt.Errorf("Ошибка при проверке файла: %v\n", err)
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func errorJSON(err error, file []byte, confname string) error {
	switch e := err.(type) {
	case *json.SyntaxError:
		line, col := offsetToLineCol(file, e.Offset)
		return fmt.Errorf("JSON syntax error in file %s, line %d, column %d: %w",
			confname, line, col, e)
	case *json.UnmarshalTypeError:
		line, col := offsetToLineCol(file, e.Offset)
		return fmt.Errorf("Type error in file %s, line %d, column %d: field %q should be %v, got %v",
			confname, line, col, e.Field, e.Type, e.Value)
	default:
		return fmt.Errorf("error reading config %s: %w", confname, err)
	}
}

func isEmptyStruct(s interface{}) bool {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Struct {
		return false
	}
	// Создаем нулевое значение того же типа
	zero := reflect.Zero(v.Type()).Interface()
	return reflect.DeepEqual(s, zero)
}
