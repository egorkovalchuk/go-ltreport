package reportdata

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"
)

// Config configuration stucture
type Config struct {
	//Имя файла
	ReportFilename string `json:"ReportFilename"`
	//Маска даты
	ReportMask string `json:"ReportMask"`
	//Включение выкладнки на конфлюенс
	ReportConfluenceOn bool `json:"ReportConfluenceOn"`
	//Адрес конфлюенса
	ReportConfluenceURL string `json:"ReportConfluenceURL"`
	//Ид куда пишем данные
	ReportConfluenceId string `json:"ReportConfluenceId"`
	//Спейс конфлюенса
	ReportConfluenceSpace string `json:"ReportConfluenceSpace"`
	ReportConfluenceLogin string `json:"ReportConfluenceLogin"`
	ReportConfluencePass  string `json:"ReportConfluencePass"`
	ReportConfluenceToken string `json:"ReportConfluenceToken"`
	ReportConfluenceProxy string `json:"ReportConfluenceProxy,omitempty"`
	ReportIM              struct {
		User       string `json:"User"`
		Pass       string `json:"Pass"`
		Token      string `json:"Token"`
		Type       int    `json:"Type"`
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
		Threshold int `json:"Threshold"`
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
		Threshold int `json:"Threshold"`
		//Описание порога
		Description string `json:"Description"`
	} `json:"JmeterQueryScnrThreshold"`
	Grafanadash []struct {
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
		Threshold int `json:"Threshold"`
		//Описание порога
		ThDescription string `json:"ThDescription"`
		//ссылка на запрос данных, смотреть в графане
		UrlQuery string `json:"UrlQuery"`
		//группировка
		UrlQueryGroup string `json:"UrlQueryGroup"`
		Size          struct {
			Width  int
			Height int
		}
	} `json:"GrafanaDash"`
	GrafanadashTemplate []struct {
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
		Threshold int `json:"Threshold"`
		// Описание порога
		ThDescription string `json:"ThDescription"`
		// ссылка на запрос данных, смотреть в графане
		UrlQuery string `json:"UrlQuery"`
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
	} `json:"GrafanadashTemplate"`
}

// Вынесена структар из конфига
type JmeterQScnrFieldS struct {
	//имя поля
	Name string `json:"Name"`
	//Описание
	Description string `json:"Description"`
}

//для хранения ключей и создания карты по порогам и их описания
//для сценариев, отличие в статусе сценария (Statut)
type KeyField struct {
	//порог
	Value int
	//Описания порога
	Description string
	//Статус сценария на Jmeter
	Statut string
}

//сруктура ошибок для анализа
type LTError struct {
	Name        string
	Threshold   int
	Description string
	Type        string
}

//Структура сценария
//Устарело, смотри ScenarioDinamic
type Scenario struct {
	Tags       string
	Percentile float64
	Maxlatency float64
	Rate       float64
	RateError  float64
}

//Структура с сценариев динамическим запросом
type ScenarioDinamic struct {
	NameTest   string
	NameThread string
	Field      []YField
}

//сруктура для вывода графиков
type LTGrag struct {
	Name        string
	Threshold   int
	Description string
	ContentType string
	UrlDash     string
	Size        struct {
		Width  int
		Height int
	}
}

// Формирование списка для динамических шаблонов
type DinamicRecord map[string]string

func BeginningOfDay() time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, time.Local)
}

// EndOfDay end of day
func EndOfDay() time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), 17, 59, 59, int(time.Second-time.Nanosecond), time.Local)
}

func BeginningOfHour() time.Time {
	t := time.Now()
	now := t.Hour()
	return time.Date(t.Year(), t.Month(), t.Day(), now-1, 0, 0, 0, time.Local)
}

// EndOfDay end of day
func EndOfHour() time.Time {
	t := time.Now()
	now := t.Hour()
	return time.Date(t.Year(), t.Month(), t.Day(), now-1, 59, 59, int(time.Second-time.Nanosecond), time.Local)
}

func AddMap(m map[string]map[string]KeyField, TestName, ErrorField string, val KeyField) map[string]map[string]KeyField {
	mm, ok := m[TestName]
	if !ok {
		mm = make(map[string]KeyField)
		m[TestName] = mm
	}
	mm[ErrorField] = val

	return m
}

func AddMapS(m map[string]map[string]ScenarioDinamic, TestName, NameThread string, val ScenarioDinamic) map[string]map[string]ScenarioDinamic {
	mm, ok := m[TestName]
	if !ok {
		mm = make(map[string]ScenarioDinamic)
		m[TestName] = mm
	}
	mm[NameThread] = val

	return m
}

//Методы для типа ScenarioDinamic
func (p *ScenarioDinamic) SetApplication(NameTest string) {
	p.NameTest = NameTest
}
func (p *ScenarioDinamic) SetThread(NameThread string) {
	p.NameThread = NameThread
}

func (p *ScenarioDinamic) SeField(YF []YField) {
	for _, i := range YF {
		p.Field = append(p.Field, i)
	}
}

func Helpstart() {
	fmt.Println("Use -v get version")
	fmt.Println("Use -d start with debug mode")
	fmt.Println("Use -c start with users config")
	fmt.Println("Use -hour to generate an hourly report ")
	fmt.Println("Use -fsmlogin start with Login FSM")
	fmt.Println("Use -fsmpass start with Password FSM")
	fmt.Println("Use -conflproxy start with proxy for connection to Confluence, example http://user:password@url:port")
	fmt.Println("Use -start and -end for generate a report on an arbitrary date ")
}

func MaxInt(x int, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func MaxInt64(x int64, y int64) int64 {
	if x > y {
		return x
	} else {
		return y
	}
}

func MinInt64(x int64, y int64) int64 {
	if x < y {
		return x
	} else {
		return y
	}
}

func ConvJsonNumFloat64(p interface{}) (float64, bool) {
	num, ok := p.(json.Number)
	if !ok {
		return 0, false
	}

	tmp, err := num.Float64()
	if err != nil {
		return 0, false
	}
	return tmp, true
}

func ConvJsonNumInt64(p interface{}) (int64, bool) {
	num, ok := p.(json.Number)
	if !ok {
		return 0, false
	}

	tmp, err := num.Int64()
	if err != nil {
		return 0, false
	}
	return tmp, true
}

func CheckStatusCode(StatusCode int, Status string) error {
	switch StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent:
		return nil
	case http.StatusNoContent, http.StatusResetContent:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("service is not available: %s", Status)
	case http.StatusInternalServerError:
		return fmt.Errorf("internal server error: %s", Status)
	case http.StatusConflict:
		return fmt.Errorf("conflict: %s", Status)
	default:
		return fmt.Errorf("unknown response status: %s", Status)
	}
}

// CalculatePercentile вычисляет заданный персентиль для набора значений
func CalculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	// Сортируем значения
	sort.Float64s(values)

	// Вычисляем индекс персентиля
	index := (percentile / 100) * float64(len(values)-1)

	// Если индекс целый - возвращаем соответствующее значение
	if index == float64(int(index)) {
		return values[int(index)]
	}

	// Интерполируем между соседними значениями
	i := int(index)
	fraction := index - float64(i)
	return values[i] + fraction*(values[i+1]-values[i])
}
