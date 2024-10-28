package confluence

import (
	"net/http"
	"net/url"
)

// API is the main api data structure
type API struct {
	Url                *url.URL
	client             *http.Client
	username, password string
	token              string
	debug              bool
}

//структура запроса (взято с доки)
type ContentQuery struct {
	Expand     []string
	Limit      int    // page limit
	OrderBy    string // fieldpath asc/desc e.g: "history.createdDate desc"
	PostingDay string // required for blogpost type Format: yyyy-mm-dd
	SpaceKey   string
	Start      int    // page start
	Status     string // current, trashed, draft, any
	Title      string // required for page
	Trigger    string // viewed
	Type       string // page, blogpost
	Version    int    //version number when not lastest

}

//json-to-go
//Сформирован автоматически
type ConflType struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Status   string   `json:"status"`
	Title    string   `json:"title"`
	Space    Space    `json:"space"`
	History  History  `json:"history"`
	Children Children `json:"children"`
	Body     struct {
		Storage    Storage  `json:"storage"`
		View       *Storage `json:"view"`
		Expandable struct {
			Editor              string `json:"editor"`
			ExportView          string `json:"export_view"`
			StyledView          string `json:"styled_view"`
			Storage             string `json:"storage"`
			AnonymousExportView string `json:"anonymous_export_view"`
		} `json:"_expandable"`
	} `json:"body"`
	Metadata   *Metadata `json:"metadata"`
	Extensions struct {
		Position interface{} `json:"position"`
	} `json:"extensions"`
	Links      *Links      `json:"_links,omitempty"`
	Expandable *Expandable `json:"_expandable"`
}

//json-to-go
//Стурктура для создания страницы
type ConflCreateType struct {
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Space     Space      `json:"space"`
	Ancestors []Ancestor `json:"ancestors,omitempty"`
	Body      Body       `json:"body"`
	Version   *Version   `json:"version,omitempty"`
}

// Space holds the Space information of a Content Page
type Space struct {
	ID     int    `json:"id,omitempty"`
	Key    string `json:"key,omitempty"`
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

// Storage defines the storage information
type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// Ancestor defines ancestors to create sub pages
type Ancestor struct {
	ID string `json:"id"`
}

// Body holds the storage information
type Body struct {
	Storage Storage  `json:"storage"`
	View    *Storage `json:"view,omitempty"`
}

// Version defines the content version number
type Version struct {
	Number    int    `json:"number"`
	MinorEdit bool   `json:"minorEdit"`
	Message   string `json:"message,omitempty"`
}

//Стуктура поиска
type Search struct {
	//Results   []Results `json:"results"`
	Start     int    `json:"start,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Size      int    `json:"size,omitempty"`
	ID        string `json:"id,omitempty"`
	TotalSize int    `json:"totalSize,omitempty"`
}

// Links contains link information
type Links struct {
	Webui      string `json:"webui"`
	Edit       string `json:"edit"`
	Tinyui     string `json:"tinyui"`
	Collection string `json:"collection"`
	Base       string `json:"base"`
	Context    string `json:"context"`
	Self       string `json:"self"`
}

// History contains object history information
type History struct {
	LastUpdated LastUpdated `json:"lastUpdated"`
	Latest      bool        `json:"latest"`
	CreatedBy   User        `json:"createdBy"`
	CreatedDate string      `json:"createdDate"`
	Links       *Links      `json:"_links"`
	Expandable  struct {
		Lastupdated     string `json:"lastUpdated"`
		Previousversion string `json:"previousVersion"`
		Contributors    string `json:"contributors"`
		Nextversion     string `json:"nextVersion"`
	}
}

// LastUpdated  contains information about the last update
type LastUpdated struct {
	By           User   `json:"by"`
	When         string `json:"when"`
	FriendlyWhen string `json:"friendlyWhen"`
	Message      string `json:"message"`
	Number       int    `json:"number"`
	MinorEdit    bool   `json:"minorEdit"`
	SyncRev      string `json:"syncRev"`
	ConfRev      string `json:"confRev"`
}

// User defines user informations
type User struct {
	Type        string `json:"type"`
	Username    string `json:"username"`
	UserKey     string `json:"userKey"`
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	Links       *Links `json:"_links"`
	Expandable  struct {
		Status string `json:"status"`
	} `json:"_expandable"`
	Profilepicture struct {
		Path      string `json:"path"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Isdefault bool   `json:"isDefault"`
	} `json:"profilePicture"`
}

type Children struct {
	Page struct {
		Results []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Status     string `json:"status"`
			Title      string `json:"title"`
			Extensions struct {
				Position interface{} `json:"position"`
			} `json:"extensions"`
			Links      *Links `json:"_links"`
			Expandable struct {
				Container    string `json:"container"`
				Metadata     string `json:"metadata"`
				Operations   string `json:"operations"`
				Children     string `json:"children"`
				Restrictions string `json:"restrictions"`
				History      string `json:"history"`
				Ancestors    string `json:"ancestors"`
				Body         string `json:"body"`
				Version      string `json:"version"`
				Descendants  string `json:"descendants"`
				Space        string `json:"space"`
			} `json:"_expandable"`
		} `json:"results"`
		Start int    `json:"start"`
		Limit int    `json:"limit"`
		Size  int    `json:"size"`
		Links *Links `json:"_links"`
	} `json:"page"`
	Links      *Links `json:"_links"`
	Expandable struct {
		Attachment string `json:"attachment"`
		Comment    string `json:"comment"`
	} `json:"_expandable"`
}

type Expandable struct {
	Container    string `json:"container"`
	Operations   string `json:"operations"`
	Restrictions string `json:"restrictions"`
	Ancestors    string `json:"ancestors"`
	Version      string `json:"version"`
	Descendants  string `json:"descendants"`
}

type Metadata struct {
	Labels struct {
		Results []interface{} `json:"results"`
		Start   int           `json:"start"`
		Limit   int           `json:"limit"`
		Size    int           `json:"size"`
		Links   *Links        `json:"_links"`
	} `json:"labels"`
	Expandable struct {
		Currentuser string `json:"currentuser"`
		Properties  string `json:"properties"`
		Frontend    string `json:"frontend"`
		Editorhtml  string `json:"editorHtml"`
	} `json:"_expandable"`
}

type ConflTypeA struct {
	Results []struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Status     string    `json:"status"`
		Title      string    `json:"title"`
		Metadata   *Metadata `json:"metadata"`
		Extensions struct {
			Mediatype string `json:"mediaType"`
			Filesize  int    `json:"fileSize"`
			Comment   string `json:"comment"`
		} `json:"extensions"`
		Links      *Links `json:"_links"`
		Expandable struct {
			Container    string `json:"container"`
			Operations   string `json:"operations"`
			Children     string `json:"children"`
			Restrictions string `json:"restrictions"`
			History      string `json:"history"`
			Ancestors    string `json:"ancestors"`
			Body         string `json:"body"`
			Version      string `json:"version"`
			Descendants  string `json:"descendants"`
			Space        string `json:"space"`
		} `json:"_expandable"`
	} `json:"results"`
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Size  int    `json:"size"`
	Links *Links `json:"_links"`
}
