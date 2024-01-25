package main

import (
	"bytes"
	"net/http"

	"github.com/Gearbox-protocol/sdk-go/log"
	"github.com/Gearbox-protocol/sdk-go/utils"
	"github.com/joho/godotenv"
)

type Config struct {
	log.CommonEnvs
	Port string `env:"PORT" default:"10101"`
}

func NewConfig() *Config {
	filename := ".env"
	godotenv.Load(filename)
	cfg := Config{}
	utils.ReadFromEnv(&cfg.CommonEnvs)
	utils.ReadFromEnv(&cfg)
	return &cfg
}

func main() {
	cfg := NewConfig()
	log.InitLogging("http server", 5, cfg.CommonEnvs)
	mux := http.NewServeMux()
	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		b := bytes.NewBuffer(nil)
		b.ReadFrom(r.Body)
		log.Info(r.Method, b.String())
		utils.HTTPServerWriteSuccess(w, "OK")
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		utils.HTTPServerWriteSuccess(w, "OK")
	})

	utils.ServerFromMux(mux, cfg.Port)
	<-make(chan int)
}
