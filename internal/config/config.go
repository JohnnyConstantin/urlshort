package config

import "flag"

var (
	AppName   = "shortener" // В дальнейшем может использоваться для CLI интерфейса
	PathToENV = ".env"      //Должен использоваться для подключения к БД
	Version   = "0.0.0.1-local"
)

var Options struct {
	Address     string
	BaseAddress string
	FileToWrite string
}

func init() {
	flag.StringVar(
		&Options.Address,
		"a",
		"localhost:8080",
		"The address to start the server on")
	flag.StringVar(
		&Options.BaseAddress,
		"b",
		"http://localhost:8080",
		"The address to return after shortener")
	flag.StringVar( // Странно, что локальные тесты намекают на необходимость этого флага в 6 инкременте, а реально появится он только в 9
		&Options.FileToWrite,
		"f",
		"log.log",
		"File to write logs",
	)
}
