package multiminer

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// Server exposes a minimal REST API to manage multiple miners.
// NOTE: This is optional for library consumers. You can use the Manager
// directly without HTTP endpoints if you're integrating into an existing application.
type Server struct {
	mgr              *Manager
	addressValidator *AddressValidator
	commandValidator *CommandValidator
}

func NewServer(mgr *Manager) *Server {
	return &Server{
		mgr:              mgr,
		addressValidator: NewAddressValidator(),
		commandValidator: NewCommandValidator(),
	}
}

func (s *Server) routes(mux *http.ServeMux) {
	// Legacy routes
	s.mountRoutes(mux, "")
	// Versioned routes
	s.mountRoutes(mux, "/api/v1")
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list := s.mgr.DeviceInfos()
		writeJSON(w, http.StatusOK, list)
	case http.MethodPost:
		var req struct {
			ID      string `json:"id"`
			Address string `json:"address"`
			Driver  string `json:"driver,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeMultiMinerError(w, NewInvalidInputError("invalid json"))
			return
		}
		if req.ID == "" || req.Address == "" {
			writeMultiMinerError(w, NewInvalidInputError("id and address are required"))
			return
		}

		// Validate address for security
		if err := s.addressValidator.ValidateAddress(req.Address); err != nil {
			writeMultiMinerError(w, err.(*MultiMinerError))
			return
		}
		var d Driver
		if req.Driver != "" {
			d = s.mgr.reg.Get(req.Driver)
			if d == nil {
				writeErrJSON(w, http.StatusBadRequest, "unknown driver")
				return
			}
		}
		if err := s.mgr.AddOrDetect(r.Context(), MinerID(req.ID), Endpoint{Address: req.Address}, d); err != nil {
			writeErrJSON(w, http.StatusBadGateway, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	default:
		writeErrJSON(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleDevice routes device-specific actions under /devices/{id}/...
func (s *Server) handleDevice(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/devices/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeErrJSON(w, http.StatusNotFound, "not found")
		return
	}
	id := MinerID(parts[0])

	if len(parts) == 1 {
		// Return basic info for the device
		writeJSON(w, http.StatusOK, map[string]interface{}{"id": id})
		return
	}

	action := parts[1]
	ctx := r.Context()
	switch r.Method + " " + action {
	case "GET summary":
		err := s.mgr.WithSession(ctx, id, func(sess Session) error {
			sm, err := sess.Summary(ctx)
			if err != nil {
				return err
			}
			writeJSON(w, http.StatusOK, sm)
			return nil
		})
		if err != nil {
			writeErrJSON(w, http.StatusBadGateway, err.Error())
		}
	case "GET stats":
		err := s.mgr.WithSession(ctx, id, func(sess Session) error {
			st, err := sess.Stats(ctx)
			if err != nil {
				return err
			}
			writeJSON(w, http.StatusOK, st)
			return nil
		})
		if err != nil {
			writeErrJSON(w, http.StatusBadGateway, err.Error())
		}
	case "POST exec":
		var req struct {
			Command   string `json:"command"`
			Parameter string `json:"parameter"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeMultiMinerError(w, NewInvalidInputError("invalid json"))
			return
		}

		// Validate command for security
		if err := s.commandValidator.ValidateCommand(req.Command); err != nil {
			writeMultiMinerError(w, err.(*MultiMinerError))
			return
		}

		if err := s.commandValidator.ValidateParameter(req.Command, req.Parameter); err != nil {
			writeMultiMinerError(w, err.(*MultiMinerError))
			return
		}
		err := s.mgr.WithSession(ctx, id, func(sess Session) error {
			data, err := sess.Exec(ctx, req.Command, req.Parameter)
			if err != nil {
				return err
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return nil
		})
		if err != nil {
			writeErrJSON(w, http.StatusBadGateway, err.Error())
		}
	case "GET power":
		err := s.mgr.WithSession(ctx, id, func(sess Session) error {
			pm, err := sess.GetPowerMode(ctx)
			if err != nil {
				return err
			}
			writeJSON(w, http.StatusOK, pm)
			return nil
		})
		if err != nil {
			writeErrJSON(w, http.StatusNotImplemented, err.Error())
		}
	case "POST power":
		var req PowerMode
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrJSON(w, http.StatusBadRequest, "invalid json")
			return
		}
		err := s.mgr.WithSession(ctx, id, func(sess Session) error {
			return sess.SetPowerMode(ctx, req)
		})
		if err != nil {
			writeErrJSON(w, http.StatusNotImplemented, err.Error())
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	case "GET fan":
		err := s.mgr.WithSession(ctx, id, func(sess Session) error {
			fc, err := sess.GetFan(ctx)
			if err != nil {
				return err
			}
			writeJSON(w, http.StatusOK, fc)
			return nil
		})
		if err != nil {
			writeErrJSON(w, http.StatusNotImplemented, err.Error())
		}
	case "POST fan":
		var req FanConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrJSON(w, http.StatusBadRequest, "invalid json")
			return
		}
		err := s.mgr.WithSession(ctx, id, func(sess Session) error {
			return sess.SetFan(ctx, req)
		})
		if err != nil {
			writeErrJSON(w, http.StatusNotImplemented, err.Error())
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	case "GET capabilities":
		// Capabilities are static per driver; read from manager device entry
		s.mgr.mu.RLock()
		dev := s.mgr.dev[id]
		s.mgr.mu.RUnlock()
		if dev == nil || dev.Driver == nil {
			writeErrJSON(w, http.StatusNotFound, "device not found")
			return
		}
		writeJSON(w, http.StatusOK, dev.Driver.Capabilities())
	default:
		writeErrJSON(w, http.StatusNotFound, "not found")
	}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErrJSON(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// writeMultiMinerError writes a structured MultiMinerError response
func writeMultiMinerError(w http.ResponseWriter, err *MultiMinerError) {
	writeJSON(w, err.HTTPStatus(), err)
}

// Start starts HTTP server on provided addr and blocks.
func (s *Server) Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	s.routes(mux)
	// Health endpoints
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	return srv.ListenAndServe()
}

// mountRoutes mounts REST routes under prefix ("" or "/api/v1").
func (s *Server) mountRoutes(mux *http.ServeMux, prefix string) {
	base := strings.TrimSuffix(prefix, "/")
	mux.HandleFunc(base+"/devices", func(w http.ResponseWriter, r *http.Request) {
		if base != "" && strings.HasPrefix(r.URL.Path, base) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, base)
		}
		s.handleDevices(w, r)
	})
	mux.HandleFunc(base+"/devices/", func(w http.ResponseWriter, r *http.Request) {
		// Strip version prefix for handler’s path parser
		if base != "" && strings.HasPrefix(r.URL.Path, base) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, base)
		}
		s.handleDevice(w, r)
	})
}
