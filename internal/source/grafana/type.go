package grafana

import (
	"net/http"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// GrafanaClient представляет клиент для работы с Grafana
type GrafanaClient struct {
	baseURL    string
	auth       string
	client     *http.Client
	logs       *logger.LogWriter
	debug      bool
	start      time.Time
	end        time.Time
	Timeperiod string
	version    int
}

// Структура алерта
type AlertV10 struct {
	ID             int       `json:"Id"`
	Version        int       `json:"Version"`
	OrgID          int       `json:"OrgId"`
	DashboardID    int       `json:"DashboardId"`
	PanelID        int       `json:"PanelId"`
	Name           string    `json:"Name"`
	Message        string    `json:"Message"`
	Severity       string    `json:"Severity"`
	State          string    `json:"State"`
	Handler        int       `json:"Handler"`
	Silenced       bool      `json:"Silenced"`
	ExecutionError string    `json:"ExecutionError"`
	Frequency      int       `json:"Frequency"`
	For            int       `json:"For"`
	EvalData       EvalData  `json:"EvalData"`
	NewStateDate   time.Time `json:"NewStateDate"`
	StateChanges   int       `json:"StateChanges"`
	Created        time.Time `json:"Created"`
	Updated        time.Time `json:"Updated"`
	Settings       struct {
		AlertRuleTags struct {
		} `json:"alertRuleTags"`
		Conditions []struct {
			Evaluator struct {
				Params []int  `json:"params"`
				Type   string `json:"type"`
			} `json:"evaluator"`
			Operator struct {
				Type string `json:"type"`
			} `json:"operator"`
			Query struct {
				DatasourceID int `json:"datasourceId"`
				Model        struct {
					Alias      string `json:"alias"`
					Datasource struct {
						Type string `json:"type"`
						UID  string `json:"uid"`
					} `json:"datasource"`
					DsType string `json:"dsType"`
					Fields []struct {
						Func string `json:"func"`
						Name string `json:"name"`
					} `json:"fields"`
					GroupBy []struct {
						Params []string `json:"params"`
						Type   string   `json:"type"`
					} `json:"groupBy"`
					GroupByTags  []interface{} `json:"groupByTags"`
					Interval     string        `json:"interval"`
					Measurement  string        `json:"measurement"`
					OrderByTime  string        `json:"orderByTime"`
					Policy       string        `json:"policy"`
					Query        string        `json:"query"`
					RefID        string        `json:"refId"`
					ResultFormat string        `json:"resultFormat"`
					Select       [][]struct {
						Params []string `json:"params"`
						Type   string   `json:"type"`
					} `json:"select"`
					Tags []struct {
						Key      string `json:"key"`
						Operator string `json:"operator"`
						Value    string `json:"value"`
					} `json:"tags"`
				} `json:"model"`
				Params []string `json:"params"`
			} `json:"query"`
			Reducer struct {
				Params []interface{} `json:"params"`
				Type   string        `json:"type"`
			} `json:"reducer"`
			Type string `json:"type"`
		} `json:"conditions"`
		ExecutionErrorState string `json:"executionErrorState"`
		For                 string `json:"for"`
		Frequency           string `json:"frequency"`
		Handler             int    `json:"handler"`
		Name                string `json:"name"`
		NoDataState         string `json:"noDataState"`
		Notifications       []struct {
			UID string `json:"uid"`
		} `json:"notifications"`
	} `json:"Settings"`
}

// Структура алерта
type AlertV11 struct {
	ID           int64   `json:"id"`
	Uid          string  `json:"uid"`
	OrgID        int64   `json:"orgID"`
	FolderUID    string  `json:"folderUID"`
	RuleGroup    string  `json:"ruleGroup"`
	Title        string  `json:"title"`
	Condition    string  `json:"condition"`
	Data         []Datum `json:"data"`
	Updated      string  `json:"updated"`
	NoDataState  string  `json:"noDataState"`
	ExecErrState string  `json:"execErrState"`
	For          string  `json:"for"`
	Annotations  struct {
		AlertID      string `json:"__alertId__"`
		DashboardUid string `json:"__dashboardUid__"`
		PanelID      string `json:"__panelId__"`
	} `json:"annotations"`
	Labels struct {
		LegacyCTamtamWfNotify string `json:"__legacy_c_tamtam_wf_notify__"`
		LegacyUseChannels     string `json:"__legacy_use_channels__"`
	} `json:"labels"`
	IsPaused             bool        `json:"isPaused"`
	NotificationSettings interface{} `json:"notification_settings"`
	Record               interface{} `json:"record"`
}

type Datum struct {
	RefID             string `json:"refId"`
	QueryType         string `json:"queryType"`
	RelativeTimeRange struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	} `json:"relativeTimeRange"`
	DatasourceUid string `json:"datasourceUid"`
	Model         any    `json:"model"`
}

type requestAnno struct {
	DashID  string
	PanelID string
	DashUID string
	State   string
}

type Annotations []Annotation

type Annotation struct {
	ID           int           `json:"id"`
	AlertID      int           `json:"alertId"`
	AlertName    string        `json:"alertName"`
	DashboardID  int           `json:"dashboardId"`
	DashboardUID string        `json:"dashboardUID"`
	PanelID      int           `json:"panelId"`
	UserID       int           `json:"userId"`
	NewState     string        `json:"newState"`
	PrevState    string        `json:"prevState"`
	Created      int64         `json:"created"`
	Updated      int64         `json:"updated"`
	Time         int64         `json:"time"`
	TimeEnd      int64         `json:"timeEnd"`
	Text         string        `json:"text"`
	Tags         []interface{} `json:"tags"`
	Login        string        `json:"login"`
	Email        string        `json:"email"`
	AvatarURL    string        `json:"avatarUrl"`
	Data         struct {
		EData EvalData
	} `json:"data"`
}

type EvalData struct {
	EvalMatches []struct {
		Metric string      `json:"metric"`
		Tags   interface{} `json:"tags"`
		Value  float64     `json:"value"`
	} `json:"evalMatches"`
}

type health struct {
	Commit  string `json:"commit"`
	Status  string `json:"comdatabasemit"`
	Version string `json:"version"`
}
