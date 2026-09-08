package influx

import (
	"net/http"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// структура influx type 1
type Mean struct {
	Results []struct {
		StatementID int `json:"statement_id"`
		Series      []struct {
			Name string `json:"name"`
			Tags struct {
				Transaction string `json:"transaction"`
				Suite       string `json:"suite"`
				Statut      string `json:"statut"`
				Application string `json:"application"`
			} `json:"tags"`
			Columns []string        `json:"columns"`
			Values  [][]interface{} `json:"values"`
		} `json:"series"`
	} `json:"results"`
}

// структура influx type 2 для сценария
type MeanScenario struct {
	Results []struct {
		StatementID int `json:"statement_id"`
		Series      []struct {
			Name string `json:"name"`
			Tags struct {
				Transaction string `json:"transaction"`
				Suite       string `json:"suite"`
				Statut      string `json:"statut"`
				Application string `json:"application"`
			} `json:"tags"`
			Columns []string        `json:"columns"`
			Values  [][]interface{} `json:"values"`
		} `json:"series"`
	} `json:"results"`
}

// для преобразования типа ответа инфлюкса
type SField struct {
	NameCol   string
	ValFloat  float64
	ValInt    int64
	ValString string
	ValTime   int64
}

// Структруа ответа инфлюкса по тестам
type LTTestDinamic struct {
	NameTest string
	Field    []YField
}
type LTTestDinamics []LTTestDinamic

// для хранения поля ответа
type YField struct {
	Name        string
	Value       float64
	Description string
	Statut      string
}

// InfluxClient представляет клиент для работы
type InfluxClient struct {
	baseURL    string
	auth       string
	client     *http.Client
	logs       *logger.LogWriter
	debug      bool
	start      time.Time
	end        time.Time
	Timeperiod string
}

// Структура с сценариев динамическим запросом
type ScenarioDinamic struct {
	NameTest   string
	NameThread string
	Field      []YField
}

type ScenarioDinamics []ScenarioDinamic

// для хранения ключей и создания карты по порогам и их описания
// для сценариев, отличие в статусе сценария (Statut)
type KeyField struct {
	//порог
	Value float64
	//Описания порога
	Description string
	//Статус сценария на Jmeter
	Statut string
}
