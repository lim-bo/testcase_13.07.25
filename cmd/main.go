package main

import (
	"log"
	"testcase/internal/api"
	archives "testcase/internal/archive_repository"
	"testcase/internal/settings"
)

func main() {
	cfg := settings.GetConfig()

	archiveManager := archives.New(cfg.GetInt("tasks_max_count"), cfg.GetInt("files_max_count"))

	serv := api.New(archiveManager)
	log.Fatal(serv.Run(cfg.GetString("api_address")))
}
