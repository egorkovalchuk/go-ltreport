package telegraf

import (
	"net/http"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// TelegrafClient представляет клиент для работы
type TelegrafClient struct {
	baseURL string
	auth    string
	client  *http.Client
	logFunc *logger.LogWriter
	debug   bool
}
