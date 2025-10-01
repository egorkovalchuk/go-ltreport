package allure

type LogStruct struct {
	T    string
	Text interface{}
}

func (a *Allure) GetError() <-chan *LogStruct {
	return a.e
}

// Запись в лог
func (a *Allure) ProcessLog(level string, logtext interface{}) {
	a.e <- &LogStruct{level, logtext}
}

// Запись в лог WARM
func (a *Allure) ProcessWarm(logtext interface{}) {
	a.e <- &LogStruct{"WARM", logtext}
}

// Запись в лог INFO
func (a *Allure) ProcessInfo(logtext interface{}) {
	a.e <- &LogStruct{"INFO", logtext}
}

// Запись в лог при включенном дебаге
func (a *Allure) ProcessDebug(logtext interface{}) {
	a.e <- &LogStruct{"DEBUG", logtext}
}

// Запись в лог ошибок
func (a *Allure) ProcessError(logtext interface{}) {
	a.e <- &LogStruct{"ERROR", logtext}
}
