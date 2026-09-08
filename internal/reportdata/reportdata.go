package reportdata

import (
	"encoding/json"
	"fmt"
	"os"
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
		ReportAllure   bool   `json:"ReportAllure"`
		ReportEndPoint string `json:"ReportEndPoint"`
		Token          string `json:"Token"`
		ReportPath     string `json:"ReportPath"`
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
	Jmeter struct {
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
	} `json:"Jmeter"`
	Grafanadash         []GrafanaDashStruct         `json:"GrafanaDash"`
	GrafanadashTemplate []GrafanadashTemplateStruct `json:"GrafanadashTemplate"`
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
	ltp := LTError{Name: "Grafana: " + name, Threshold: threshold, Description: description + ": Threshold " + fmt.Sprint(threshold) + " - current " + strconv.Itoa(int(percentile)), Type: tp, Tag: tag}
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
	file, err := os.Open(confname)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println(err)
		return err
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
	file.Close()
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
