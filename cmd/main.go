package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Gearbox-protocol/sdk-go/log"
	"github.com/Gearbox-protocol/sdk-go/utils"
	"github.com/joho/godotenv"
)

type Config struct {
	log.CommonEnvs
	Port      string `env:"PORT" default:"10101"`
	LogPath   string `env:"LOG_PATH"`
	AllowedMB int64  `env:"ALLOWED_MB" default:"20"`
}

func NewConfig() *Config {
	filename := ".env"
	godotenv.Load(filename)
	cfg := Config{}
	utils.ReadFromEnv(&cfg.CommonEnvs)
	utils.ReadFromEnv(&cfg)
	return &cfg
}

type FlyConfig struct {
	App    map[string]string `json:"app"`
	Region string            `json:"region"`
}
type Log struct {
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
	Fly       FlyConfig `json:"fly"`
}

func main() {
	cfg := NewConfig()
	log.InitLogging("http server", 5, cfg.CommonEnvs)
	mux := http.NewServeMux()
	mux.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		logs := &[]Log{}
		err := utils.ReadJsonReaderAndSetInterface(r.Body, logs)
		log.CheckFatal(err)
		for _, entry := range *logs {
			saveLog(entry, cfg.LogPath, cfg.AllowedMB)
		}
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		utils.HTTPServerWriteSuccess(w, "OK")
	})

	utils.ServerFromMux(mux, cfg.Port)
	<-make(chan int)
}

func saveLog(l Log, path string, allowedMB int64) {
	fileName := getfileName(l)
	save(filepath.Join(path, fileName), l.Message, allowedMB)

}

func save(fileUri string, message string, allowedMB int64) {
	file, err := os.OpenFile(fileUri, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.CheckFatal(err)
	fi, err := file.Stat()
	if err != nil {
		log.Error("can't write file", err)
	}
	if fi.Size()/1024*1024 >= allowedMB {
		mid := allowedMB * 1024 * 1024 / 2
		file.Seek(mid, 0)
		file.Truncate(mid)
	}
	_, err = file.WriteString(message + "\n")
	if err != nil {
		log.Error("can't write file", err)
	}
}
func getfileName(l Log) string {
	return fmt.Sprintf("%s-%s-%s.log", l.Fly.App["name"], l.Fly.App["instance"], l.Fly.Region)
}
