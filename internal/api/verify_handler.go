package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Upellift99/GateCHA/internal/altcha"
	"github.com/Upellift99/GateCHA/internal/models"

	lib "github.com/altcha-org/altcha-lib-go"
)

type VerifyHandler struct {
	DB *sql.DB
}

type verifyRequest struct {
	Payload string `json:"payload"`
}

type verifyResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func (h *VerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := GetAPIKeyFromContext(r)
	if key == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing API key"})
		return
	}

	slog.Debug("verify request", "api_key_id", key.ID, "key_id", key.KeyID)

	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, verifyResponse{OK: false, Error: "invalid request body"})
		return
	}

	if req.Payload == "" {
		writeJSON(w, http.StatusBadRequest, verifyResponse{OK: false, Error: "missing payload"})
		return
	}

	// Decode payload to extract challenge hash for replay check
	decoded, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		if err := models.IncrementVerificationsFail(h.DB, key.ID); err != nil {
			slog.Error("failed to increment verifications_fail", "error", err, "api_key_id", key.ID)
		}
		writeJSON(w, http.StatusOK, verifyResponse{OK: false, Error: "invalid payload encoding"})
		return
	}

	var payload lib.Payload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		if err := models.IncrementVerificationsFail(h.DB, key.ID); err != nil {
			slog.Error("failed to increment verifications_fail", "error", err, "api_key_id", key.ID)
		}
		writeJSON(w, http.StatusOK, verifyResponse{OK: false, Error: "invalid payload format"})
		return
	}

	// Verify the solution
	ok, err := altcha.VerifyPayload(key.HMACSecret, req.Payload)
	if err != nil {
		if err := models.IncrementVerificationsFail(h.DB, key.ID); err != nil {
			slog.Error("failed to increment verifications_fail", "error", err, "api_key_id", key.ID)
		}
		writeJSON(w, http.StatusOK, verifyResponse{OK: false, Error: "verification failed"})
		return
	}

	if !ok {
		if err := models.IncrementVerificationsFail(h.DB, key.ID); err != nil {
			slog.Error("failed to increment verifications_fail", "error", err, "api_key_id", key.ID)
		}
		writeJSON(w, http.StatusOK, verifyResponse{OK: false, Error: "invalid_solution"})
		return
	}

	// Check replay
	consumed, err := models.IsConsumed(h.DB, payload.Challenge)
	if err != nil {
		slog.Error("failed to check consumed", "error", err, "api_key_id", key.ID)
		writeJSON(w, http.StatusInternalServerError, verifyResponse{OK: false, Error: "internal error"})
		return
	}
	if consumed {
		if err := models.IncrementVerificationsFail(h.DB, key.ID); err != nil {
			slog.Error("failed to increment verifications_fail", "error", err, "api_key_id", key.ID)
		}
		writeJSON(w, http.StatusOK, verifyResponse{OK: false, Error: "already_used"})
		return
	}

	// Mark as consumed
	expiresAt := time.Now().Add(time.Duration(key.ExpireSeconds) * time.Second)
	if err := models.MarkConsumed(h.DB, payload.Challenge, key.ID, expiresAt); err != nil {
		slog.Error("failed to mark consumed", "error", err, "api_key_id", key.ID)
	}
	if err := models.IncrementVerificationsOK(h.DB, key.ID); err != nil {
		slog.Error("failed to increment verifications_ok", "error", err, "api_key_id", key.ID)
	}

	slog.Debug("verify success", "api_key_id", key.ID)
	writeJSON(w, http.StatusOK, verifyResponse{OK: true})
}
