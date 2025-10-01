package allure

type TokenResponse struct {
	ExpiresIn   int64  `json:"expires_in"`
	Jti         string `json:"jti"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type AllureResult struct {
	UUID        string             `json:"uuid"`
	Name        string             `json:"name"`
	FullName    string             `json:"fullName"`
	HistoryID   string             `json:"historyId"`
	Status      string             `json:"status"`
	Stage       string             `json:"stage"`
	Start       int64              `json:"start"`
	Stop        int64              `json:"stop"`
	Steps       []AllureStep       `json:"steps"`
	Labels      []AllureLabel      `json:"labels"`
	Links       []AllureLink       `json:"links"`
	Attachments []AllureAttachment `json:"attachments"`
	Parameters  []AllureParameter  `json:"parameters"`
	Description string             `json:"description,omitempty"`
}

type AllureStep struct {
	Name        string             `json:"name"`
	Status      string             `json:"status"`
	Start       int64              `json:"start"`
	Stop        int64              `json:"stop"`
	Steps       []AllureStep       `json:"steps"`
	Attachments []AllureAttachment `json:"attachments"`
	Parameters  []AllureParameter  `json:"parameters"`
}

// AllureParameter - структура параметра
type AllureParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AllureAttachment struct {
	Name    string `json:"name"`
	Source  string `json:"source"`            // UUID названия файла
	Type    string `json:"type"`              // MIME type
	Content string `json:"content,omitempty"` // base64 encoded content
}

type AllureLabel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AllureLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

type AllureLaunchTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type LaunchRequest struct {
	ProjectID int               `json:"projectId"`
	Name      string            `json:"name"`
	External  bool              `json:"external"`
	Autoclose bool              `json:"autoclose"`
	Tags      []AllureLaunchTag `json:"tags"`
	Links     []AllureLink      `json:"links"`
	Issues    []struct {
		ID            int    `json:"id"`
		IntegrationID int    `json:"integrationId"`
		Name          string `json:"name"`
		URL           string `json:"url"`
		Summary       string `json:"summary"`
		Status        string `json:"status"`
		Closed        bool   `json:"closed"`
	} `json:"issues"`
}

type LaunchResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Closed    bool   `json:"closed"`
	External  bool   `json:"external"`
	Autoclose bool   `json:"autoclose"`
	ProjectID int    `json:"projectId"`
	Tags      []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
	Links  []AllureLink `json:"links"`
	Issues []struct {
		ID            int    `json:"id"`
		IntegrationID int    `json:"integrationId"`
		Name          string `json:"name"`
		URL           string `json:"url"`
		Summary       string `json:"summary"`
		Status        string `json:"status"`
		Closed        bool   `json:"closed"`
	} `json:"issues"`
	CreatedDate      int    `json:"createdDate"`
	LastModifiedDate int    `json:"lastModifiedDate"`
	CreatedBy        string `json:"createdBy"`
	LastModifiedBy   string `json:"lastModifiedBy"`
}

type LaunchRequestArchive struct {
	EnvVarValues []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Variable struct {
			ID               int    `json:"id"`
			Name             string `json:"name"`
			CreatedDate      int    `json:"createdDate"`
			LastModifiedDate int    `json:"lastModifiedDate"`
			CreatedBy        string `json:"createdBy"`
			LastModifiedBy   string `json:"lastModifiedBy"`
		} `json:"variable"`
	} `json:"envVarValues"`
}

type UploadResponse struct {
	LaunchID      int `json:"launchId"`
	TestSessionID int `json:"testSessionId"`
	FilesCount    int `json:"filesCount"`
}

type GetLaunchResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Closed    bool   `json:"closed"`
	External  bool   `json:"external"`
	Autoclose bool   `json:"autoclose"`
	ProjectID int    `json:"projectId"`
	Tags      []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
	Links  []AllureLink `json:"links"`
	Issues []struct {
		ID            int    `json:"id"`
		IntegrationID int    `json:"integrationId"`
		Name          string `json:"name"`
		URL           string `json:"url"`
		Summary       string `json:"summary"`
		Status        string `json:"status"`
		Closed        bool   `json:"closed"`
	} `json:"issues"`
	CreatedDate      int    `json:"createdDate"`
	LastModifiedDate int    `json:"lastModifiedDate"`
	CreatedBy        string `json:"createdBy"`
	LastModifiedBy   string `json:"lastModifiedBy"`
}

type LaunchStatResponse []struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
