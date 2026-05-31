package policy

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"go.uber.org/zap"
)

// RiskLevel represents the compliance risk classification
type RiskLevel string

const (
	RiskLevelL1 RiskLevel = "L1" // Green - Safe, direct routing
	RiskLevelL2 RiskLevel = "L2" // Orange - Enhanced verification
	RiskLevelL3 RiskLevel = "L3" // Red - Manual review
	RiskLevelL4 RiskLevel = "L4" // Dark red - Blocked
)

// RoutingPath represents the transaction routing path
type RoutingPath string

const (
	PathA      RoutingPath = "PATH_A" // Direct settlement (L1 only)
	PathB      RoutingPath = "PATH_B" // Enhanced checks (L2)
	PathC      RoutingPath = "PATH_C" // Manual approval (L3)
	Blocked    RoutingPath = "BLOCKED" // Sanctions list (L4)
)

// Transaction represents a transaction for policy evaluation
type Transaction struct {
	ID                string            `json:"id"`
	Amount            float64           `json:"amount"`
	Currency          string            `json:"currency"`
	SenderCountry     string            `json:"sender_country"`
	ReceiverCountry   string            `json:"receiver_country"`
	TransactionType   string            `json:"transaction_type"`
	SenderRiskScore   int               `json:"sender_risk_score"`
	ReceiverRiskScore int               `json:"receiver_risk_score"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// PolicyEvaluation represents the result of policy evaluation
type PolicyEvaluation struct {
	TransactionID  string                 `json:"transaction_id"`
	RiskLevel      RiskLevel              `json:"risk_level"`
	RoutingPath    RoutingPath            `json:"routing_path"`
	Approved       bool                   `json:"approved"`
	Reason         string                 `json:"reason"`
	RequiredReview string                 `json:"required_review,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      int64                  `json:"timestamp"`
}

// PolicyEngine manages OPA policy evaluation
type PolicyEngine struct {
	logger *zap.Logger
	query  rego.PreparedEvalQuery
}

// NewPolicyEngine creates a new policy engine instance
func NewPolicyEngine(logger *zap.Logger) (*PolicyEngine, error) {
	// Load policy from embedded OPA rules
	policyText := `
package sovereign_trade

# Risk Level Classification
classify_risk_level[level] {
	input.transaction.sender_risk_score >= 80 | input.transaction.receiver_risk_score >= 80
	level := "L4"
} {
	input.transaction.sender_risk_score >= 60 | input.transaction.receiver_risk_score >= 60
	level := "L3"
} {
	input.transaction.sender_risk_score >= 40 | input.transaction.receiver_risk_score >= 40
	level := "L2"
} {
	level := "L1"
}

# Routing Path Decision
route_transaction[path] {
	classify_risk_level["L4"]
	path := "BLOCKED"
} {
	classify_risk_level["L3"]
	path := "PATH_C"
} {
	classify_risk_level["L2"]
	path := "PATH_B"
} {
	classify_risk_level["L1"]
	path := "PATH_A"
}

# Transaction Approval
approve_transaction[approved] {
	route_transaction["BLOCKED"]
	approved := false
} {
	route_transaction[_]
	approved := true
}
`

	// Prepare the query
	query, err := rego.New(
		rego.Query("result = {\"risk\": sovereign_trade.classify_risk_level[x], \"path\": sovereign_trade.route_transaction[y], \"approved\": sovereign_trade.approve_transaction[z]}"),
		rego.Module("policy.rego", policyText),
	).PrepareForEval(context.Background())

	if err != nil {
		return nil, fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	return &PolicyEngine{
		logger: logger,
		query:  query,
	}, nil
}

// EvaluateTransaction evaluates a transaction against policies
func (pe *PolicyEngine) EvaluateTransaction(ctx context.Context, tx *Transaction) (*PolicyEvaluation, error) {
	// Prepare input for OPA
	input := map[string]interface{}{
		"transaction": tx,
	}

	// Execute policy evaluation
	results, err := pe.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		pe.logger.Error("Policy evaluation error", zap.Error(err), zap.String("tx_id", tx.ID))
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no policy results found")
	}

	// Extract results
	result := results[0].Bindings["result"].(map[string]interface{})

	riskLevel := RiskLevelL1
	if risk, ok := result["risk"].(ast.Term); ok {
		riskLevel = RiskLevel(risk.String())
	}

	routingPath := PathA
	if path, ok := result["path"].(ast.Term); ok {
		routingPath = RoutingPath(path.String())
	}

	approved := result["approved"].(bool)

	reason := pe.generateReason(tx, riskLevel, routingPath)

	evaluation := &PolicyEvaluation{
		TransactionID: tx.ID,
		RiskLevel:     riskLevel,
		RoutingPath:   routingPath,
		Approved:      approved,
		Reason:        reason,
		Metadata:      make(map[string]interface{}),
		Timestamp:     int64(time.Now().Unix()),
	}

	if routingPath == PathC {
		evaluation.RequiredReview = "MANUAL_REVIEW_REQUIRED"
	}

	pe.logger.Info("Transaction evaluated",
		zap.String("tx_id", tx.ID),
		zap.String("risk_level", string(riskLevel)),
		zap.String("routing_path", string(routingPath)),
		zap.Bool("approved", approved),
	)

	return evaluation, nil
}

func (pe *PolicyEngine) generateReason(tx *Transaction, level RiskLevel, path RoutingPath) string {
	switch level {
	case RiskLevelL1:
		return fmt.Sprintf("Low risk transaction: sender_score=%d, receiver_score=%d",
			tx.SenderRiskScore, tx.ReceiverRiskScore)
	case RiskLevelL2:
		return fmt.Sprintf("Medium risk - Enhanced verification required: sender_score=%d, receiver_score=%d",
			tx.SenderRiskScore, tx.ReceiverRiskScore)
	case RiskLevelL3:
		return fmt.Sprintf("High risk - Manual review required: sender_score=%d, receiver_score=%d, countries=[%s→%s]",
			tx.SenderRiskScore, tx.ReceiverRiskScore, tx.SenderCountry, tx.ReceiverCountry)
	case RiskLevelL4:
		return fmt.Sprintf("Critical risk - BLOCKED: Sanctions list match or prohibited jurisdiction detected")
	default:
		return "Unknown risk level"
	}
}
