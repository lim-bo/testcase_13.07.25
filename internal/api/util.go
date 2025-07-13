package api

import (
	"net/http"

	"github.com/bytedance/sonic"
)

func writeMessage(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_ = sonic.ConfigFastest.NewEncoder(w).Encode(map[string]interface{}{
		"cod": code,
		"msg": message,
	})
}
