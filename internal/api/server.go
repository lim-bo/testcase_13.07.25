package api

import (
	"log/slog"
	"net/http"
	"testcase/models"

	"github.com/go-chi/chi/v5"
)

type ArchiveManager interface {
	CreateTask() (string, error)
	AddFile(taskID string, file *models.FileRequest) error
	GetTaskStatus(taskID string) (models.TaskStatus, error)
	GetArchive(taskID string) ([]byte, error)
	SaveArchiveLocaly(raw []byte, taskID string) (string, error)
}

type Server struct {
	mx *chi.Mux
	am ArchiveManager
}

func New(am ArchiveManager) *Server {
	return &Server{
		mx: chi.NewMux(),
		am: am,
	}
}

func (s *Server) mountEndpoints() {
	s.mx.Use(s.CORSMiddleware, s.RequestIDMiddleware)
	s.mx.Route("/tasks", func(r chi.Router) {
		r.Put("/create", s.createTask)
		r.Post("/{id}/add", s.addFileToTask)
		r.Get("/{id}/check", s.getTask)
	})
	fs := http.FileServer(http.Dir("./data"))
	s.mx.Handle("/download/*", http.StripPrefix("/download", fs))
}

func (s *Server) Run(addr string) error {
	s.mountEndpoints()
	slog.Info("starting server at " + addr)
	return http.ListenAndServe(addr, s.mx)
}
