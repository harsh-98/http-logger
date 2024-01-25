package main

import (
	"fmt"
	"io"
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
		if err != nil {
			log.Error(err)
			return
		}
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

var MB int64 = 1024 * 1024

func save(fileUri string, message string, allowedMB int64) {
	MB = 1
	file, err := os.OpenFile(fileUri, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.CheckFatal(err)
	fi, err := file.Stat()
	if err != nil {
		log.Error("can't read  file stats", err)
		file.Close()
		return
	}
	if fi.Size()/MB >= allowedMB {
		mid := allowedMB / 2
		copyFile(file, mid, fileUri)
		//
		file, err = os.OpenFile(fileUri, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Error(log.WrapErrWithLine(err))
			return
		}
	}
	defer file.Close()
	_, err = file.WriteString(message + "\n")
	if err != nil {
		log.Error("can't write file", err)
		return
	}
}
func getfileName(l Log) string {
	// l.Fly.App["instance"],
	return fmt.Sprintf("%s-%s.log", l.Fly.App["name"], l.Fly.Region)
}

func copyFile(inFile *os.File, mid int64, fileName string) error {
	tmpFile := "/tmp/dest.log"
	// Offset is the number of bytes you want to exclude
	_, err := inFile.Seek(mid, io.SeekStart)
	if err != nil {
		return log.WrapErrWithLine(err)
	}

	fout, err := os.Create(tmpFile)
	if err != nil {
		return log.WrapErrWithLine(err)
	}

	_, err = io.Copy(fout, inFile)
	if err != nil {
		return log.WrapErrWithLine(err)
	}
	fout.Close()
	inFile.Close()
	err = os.Rename(tmpFile, fileName)
	if err != nil {
		return log.WrapErrWithLine(err)
	}
	return nil
}
