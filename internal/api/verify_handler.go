package api

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Upellift99/GateCHA/internal/altcha"
	"github.com/Upellift99/GateCHA/internal/geo"
	"github.com/Upellift99/GateCHA/internal/his"
	"github.com/Upellift99/GateCHA/internal/models"
	lib "github.com/altcha-org/altcha-lib-go"
	"gorm.io/gorm"
)

const logMsgFailIncrement = "failed to increment verifications_fail"

type VerifyHandler struct {
	DB *gorm.DB
}

type verifyRequest struct {
	Payload string `json:"payload"`
	// HISSignals is an optional privacy-preserving interaction sample. When
	// present it is scored and recorded. It changes the verification result
	// only on a key that has explicitly opted into HIS enforcement; otherwise
	// the score is reported and nothing else.
	HISSignals *his.Signals `json:"his_signals,omitempty"`
}

type verifyResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	// HISBotScore is the Monitor-mode interaction score in [0,1], where higher
	// means more bot-like. Mind the polarity: it is the reverse of reCAPTCHA's
	// convention, which is why the field is not called "score". It is a pointer
	// so that "no signals were sent" stays distinguishable from a legitimate
	// score of 0; a client that ships no collector is not thereby a human.
	HISBotScore *float64 `json:"his_bot_score,omitempty"`
	// HISBotSuspected is HISBotScore judged against this key's suspect
	// threshold, for callers who would rather not choose a number themselves.
	// Omitted alongside the score. On a key with enforcement on, a true here
	// arrives with OK false and error "bot_suspected".
	HISBotSuspected *bool `json:"his_bot_suspected,omitempty"`
}

func (h *VerifyHandler) recordFail(apiKeyID int64, country string) {
	if err := models.IncrementVerificationsFail(h.DB, apiKeyID); err != nil {
		slog.Error(logMsgFailIncrement, "error", err, "api_key_id", apiKeyID)
	}
	h.recordCountry(apiKeyID, country, false)
}

// recordCountry attributes one verification outcome to the source country
// (resolved at request time; the IP itself is never stored). country may be
// empty for unlocatable sources, which is bucketed as "unknown".
func (h *VerifyHandler) recordCountry(apiKeyID int64, country string, ok bool) {
	if err := models.IncrementCountryVerification(h.DB, apiKeyID, country, ok); err != nil {
		slog.Error("failed to increment country stat", "error", err, "api_key_id", apiKeyID)
	}
}

func (h *VerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := GetAPIKeyFromContext(r)
	if key == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing API key"})
		return
	}

	slog.Debug("verify request", "api_key_id", key.ID, "key_id", key.KeyID)

	// Resolve the source country once, then discard the IP. Used only for the
	// aggregated per-country breakdown; the raw address is never stored.
	country := geo.Country(clientIP(r))

	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, verifyResponse{OK: false, Error: "invalid request body"})
		return
	}

	// Score HIS before outcome branching so bot-like attempts that fail the PoW
	// are still observed. Scoring itself never affects the outcome; enforcement
	// is applied at the end of the happy path, if the key opted in.
	hisOut := recordHISMonitor(h.DB, key, req.HISSignals)

	// respond attaches the Monitor score to every reply from here on, failures
	// included: a submission can fail the PoW and still be worth scoring, and a
	// caller enforcing its own threshold wants both cases.
	respond := func(status int, res verifyResponse) {
		if hisOut != nil {
			res.HISBotScore = &hisOut.Score
			res.HISBotSuspected = &hisOut.Suspected
		}
		writeJSON(w, status, res)
	}

	if req.Payload == "" {
		respond(http.StatusBadRequest, verifyResponse{OK: false, Error: "missing payload"})
		return
	}

	// Decode payload to extract challenge hash for replay check
	decoded, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		h.recordFail(key.ID, country)
		respond(http.StatusOK, verifyResponse{OK: false, Error: "invalid payload encoding"})
		return
	}

	var payload lib.Payload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		h.recordFail(key.ID, country)
		respond(http.StatusOK, verifyResponse{OK: false, Error: "invalid payload format"})
		return
	}

	// Verify the solution
	ok, err := altcha.VerifyPayload(key.HMACSecret, req.Payload)
	if err != nil {
		h.recordFail(key.ID, country)
		respond(http.StatusOK, verifyResponse{OK: false, Error: "verification failed"})
		return
	}

	if !ok {
		h.recordFail(key.ID, country)
		respond(http.StatusOK, verifyResponse{OK: false, Error: "invalid_solution"})
		return
	}

	// Check replay
	consumed, err := models.IsConsumed(h.DB, payload.Challenge)
	if err != nil {
		slog.Error("failed to check consumed", "error", err, "api_key_id", key.ID)
		respond(http.StatusInternalServerError, verifyResponse{OK: false, Error: "internal error"})
		return
	}
	if consumed {
		h.recordFail(key.ID, country)
		respond(http.StatusOK, verifyResponse{OK: false, Error: "already_used"})
		return
	}

	// Mark as consumed
	expiresAt := time.Now().Add(time.Duration(key.ExpireSeconds) * time.Second)
	if err := models.MarkConsumed(h.DB, payload.Challenge, key.ID, expiresAt); err != nil {
		slog.Error("failed to mark consumed", "error", err, "api_key_id", key.ID)
	}
	// Enforcement, last: the proof of work has passed and the challenge is
	// consumed, so a rejected attempt cannot be retried with the same solved
	// payload and better-looking signals. Placed after MarkConsumed for that
	// reason, and after every PoW failure branch so `invalid_solution` keeps
	// precedence: a submission that failed the maths is not reported as a bot.
	//
	// hisOut is nil when no signals arrived, which is what keeps a site with no
	// collector immune to this switch entirely.
	if key.HISEnforce && hisOut != nil && hisOut.Suspected {
		slog.Info("his enforcement rejected verification",
			"api_key_id", key.ID, "score", hisOut.Score, "threshold", key.SuspectThreshold())
		h.recordFail(key.ID, country)
		respond(http.StatusOK, verifyResponse{OK: false, Error: "bot_suspected"})
		return
	}

	if err := models.IncrementVerificationsOK(h.DB, key.ID); err != nil {
		slog.Error("failed to increment verifications_ok", "error", err, "api_key_id", key.ID)
	}
	h.recordCountry(key.ID, country, true)

	slog.Debug("verify success", "api_key_id", key.ID)
	respond(http.StatusOK, verifyResponse{OK: true})
}
