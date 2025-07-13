package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	errvalues "testcase/internal/errors"
	"testcase/internal/settings"
	"testcase/models"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "Request-ID", reqID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value("Request-ID").(string)
	w.Header().Set("Content-Type", "application/json")
	taskID, err := s.am.CreateTask()
	if err != nil {
		slog.Error("creating task error", slog.String("error", err.Error()))
		writeMessage(w, http.StatusInternalServerError, "repository error")
		return
	}
	err = sonic.ConfigFastest.NewEncoder(w).Encode(map[string]any{"cod": http.StatusOK, "task_id": taskID})
	if err != nil {
		slog.Error("error marshalling results", slog.String("error", err.Error()))
		writeMessage(w, http.StatusInternalServerError, "json error")
		return
	}
	slog.Info("successfully created task", slog.String("req_id", reqID))
}

func (s *Server) addFileToTask(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value("Request-ID").(string)
	w.Header().Set("Content-Type", "application/json")
	taskID := r.PathValue("id")
	if taskID == "" {
		slog.Error("incoming request with invalid id in path-value", slog.String("req_id", reqID))
		writeMessage(w, http.StatusBadRequest, "invalid taskID in path")
		return
	}
	var file models.FileRequest
	err := sonic.ConfigFastest.NewDecoder(r.Body).Decode(&file)
	if err != nil {
		slog.Error("error unmarshalling request body", slog.String("error", err.Error()), slog.String("req_id", reqID))
		writeMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = s.am.AddFile(taskID, &file)
	if err != nil {
		if errors.Is(err, errvalues.ErrNoSuchTask) {
			slog.Error("incoming request for unexist task", slog.String("req_id", reqID))
			writeMessage(w, http.StatusBadRequest, "task with such id doesn't exist")
			return
		} else if errors.Is(err, errvalues.ErrTaskFull) {
			slog.Error("request for adding file to full task", slog.String("req_id", reqID))
			writeMessage(w, http.StatusBadRequest, "max files count for this task exceeded")
			return
		}
		slog.Error("adding file error", slog.String("error", err.Error()))
		writeMessage(w, http.StatusInternalServerError, "repository error")
		return
	}
	writeMessage(w, http.StatusOK, "file added to task")
	slog.Info("successfully added file to task", slog.String("req_id", reqID))
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value("Request-ID").(string)
	w.Header().Set("Content-Type", "application/json")
	taskID := r.PathValue("id")
	if taskID == "" {
		slog.Error("incoming request with invalid id in path-value", slog.String("req_id", reqID))
		writeMessage(w, http.StatusBadRequest, "invalid taskID in path")
		return
	}
	status, err := s.am.GetTaskStatus(taskID)
	if err != nil {
		if errors.Is(err, errvalues.ErrNoSuchTask) {
			slog.Error("incoming request for unexist task", slog.String("req_id", reqID))
			writeMessage(w, http.StatusBadRequest, "task with such id doesn't exist")
			return
		}
		slog.Error("getting status error", slog.String("req_id", reqID))
		writeMessage(w, http.StatusInternalServerError, "repository error")
		return
	}
	switch status {
	case models.Staged:
		err = sonic.ConfigFastest.NewEncoder(w).Encode(map[string]any{
			"cod":    200,
			"id":     taskID,
			"status": status,
		})
		if err != nil {
			slog.Error("marshalling results error", slog.String("error", err.Error()))
			writeMessage(w, http.StatusInternalServerError, "json error")
			return
		}
	case models.Completed:
		archiveData, err := s.am.GetArchive(taskID)
		if err != nil {
			if errors.Is(err, errvalues.ErrNoSuchTask) {
				slog.Error("incoming request for unexist task", slog.String("req_id", reqID))
				writeMessage(w, http.StatusBadRequest, "task with such id doesn't exist")
				return
			}
			slog.Error("getting archive error", slog.String("req_id", reqID), slog.String("error", err.Error()))
			writeMessage(w, http.StatusInternalServerError, "repository error")
			return
		}
		filename, err := s.am.SaveArchiveLocaly(archiveData, taskID)
		if err != nil {
			slog.Error("saving archive error", slog.String("req_id", reqID), slog.String("error", err.Error()))
			writeMessage(w, http.StatusInternalServerError, "repository error")
			return
		}
		err = sonic.ConfigFastest.NewEncoder(w).Encode(map[string]any{
			"code":   200,
			"id":     taskID,
			"link":   "http://" + settings.GetConfig().GetString("api_address") + "/download/" + filename,
			"status": models.Completed,
		})
		if err != nil {
			slog.Error("marshalling results error", slog.String("error", err.Error()))
			writeMessage(w, http.StatusInternalServerError, "json error")
			return
		}
	}
	slog.Info("succesfully provided task status", slog.String("req_id", reqID))
}
