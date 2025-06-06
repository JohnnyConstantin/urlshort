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
		"http://localhost:8080/",
		"The address to return after shortener")
}
