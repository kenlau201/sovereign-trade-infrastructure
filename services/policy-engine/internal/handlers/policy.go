package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/kenlau201/sovereign-trade-infrastructure/services/policy-engine/internal/policy"
)

// PolicyHandler handles policy-related HTTP requests
type PolicyHandler struct {
	engine *policy.PolicyEngine
	logger *zap.Logger
}

// NewPolicyHandler creates a new policy handler
func NewPolicyHandler(engine *policy.PolicyEngine, logger *zap.Logger) *PolicyHandler {
	return &PolicyHandler{
		engine: engine,
		logger: logger,
	}
}

// EvaluatePolicy evaluates a transaction against compliance policies
func (ph *PolicyHandler) EvaluatePolicy(w http.ResponseWriter, r *http.Request) {
	var tx policy.Transaction

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		ph.logger.Error("Request parsing error", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Generate transaction ID if not provided
	if tx.ID == "" {
		tx.ID = uuid.New().String()
	}

	// Evaluate policy
	evaluation, err := ph.engine.EvaluateTransaction(r.Context(), &tx)
	if err != nil {
		ph.logger.Error("Policy evaluation error", zap.Error(err))
		http.Error(w, "Policy evaluation failed", http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(evaluation)
}

// ListPolicies lists available policies
func (ph *PolicyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies := map[string]interface{}{
		"policies": []map[string]interface{}{
			{
				"name":        "risk_classification",
				"description": "Classify transaction risk level (L1-L4)",
				"version":     "1.0.0",
			},
			{
				"name":        "routing_decision",
				"description": "Determine routing path (PATH_A/B/C or BLOCKED)",
				"version":     "1.0.0",
			},
			{
				"name":        "approval_logic",
				"description": "Determine if transaction is approved",
				"version":     "1.0.0",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(policies)
}

// ReloadPolicies reloads policies (for future use)
func (ph *PolicyHandler) ReloadPolicies(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "success",
		"message":   "Policies reloaded",
		"timestamp": int64(time.Now().Unix()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
