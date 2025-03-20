package services

import (
	"context"
	"cudo-test/gen"
	"cudo-test/internal/repos"
	"fmt"
	"math"
	"sync"
	"time"
)

type MainServiceImpl struct {
	mainRepo repos.MainRepo
}

type FraudResult struct {
	TransactionID    string                 `json:"transaction_id"`
	FraudScore       float64                `json:"fraud_score"`
	RiskLevel        string                 `json:"risk_level"`
	DetectionResults map[string]interface{} `json:"detection_results"`
}

type BatchResult struct {
	Transactions   []FraudResult          `json:"transactions"`
	ProcessingMeta map[string]interface{} `json:"processing_metadata"`
}

type CheckResult struct {
	Result   map[string]interface{}
	Duration time.Duration
}

type CheckDurations struct {
	FreqDuration    time.Duration
	AmountDuration  time.Duration
	PatternDuration time.Duration
}

type ProcessingResult struct {
	FraudResult FraudResult
	Durations   CheckDurations
}

type MainService interface {
	DetectFraud(limit int32, riskLevels []string) (BatchResult, error)
}

func InitMainService(mainRepo repos.MainRepo) *MainServiceImpl {
	return &MainServiceImpl{mainRepo: mainRepo}
}

const (
	HighRiskScoreThreshold       = 80.0
	MediumRiskScoreThreshold     = 50.0
	FrequencyHighThreshold       = 8
	FrequencyMediumHighThreshold = 7
	FrequencyMediumThreshold     = 6
	FrequencyLowThreshold        = 5
	PatternSpikeThreshold        = 300.0 // percentage increase
)

// DetectFraud analyzes recent transactions for potential fraud.
//
// It fetches recent transactions, groups them by user, and processes each transaction
// in parallel using goroutines.
//
// For each transaction, it performs three checks:
// frequency, amount, and pattern analysis. The results are combined into a fraud score,
// and transactions are filtered by specified risk levels. Processing metadata, including
// average check durations, is included in the response.
func (s *MainServiceImpl) DetectFraud(limit int32, riskLevels []string) (BatchResult, error) {
	startTime := time.Now()

	// Fetch recent transactions
	transactions, err := s.mainRepo.GetRecentTransactions(context.Background(), limit)
	if err != nil {
		return BatchResult{}, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	// Group transactions by user ID for efficient lookup during checks
	userTransactionsMap := make(map[int64][]gen.Transaction)
	for _, tx := range transactions {
		userTransactionsMap[tx.UserID] = append(userTransactionsMap[tx.UserID], tx)
	}

	var wg sync.WaitGroup
	processingResultsChan := make(chan ProcessingResult, len(transactions))

	// Process each transaction in parallel
	for _, tx := range transactions {
		wg.Add(1)
		go func(tx gen.Transaction) {
			defer wg.Done()

			var checkWg sync.WaitGroup
			checkWg.Add(3)

			// Variables to store results from parallel checks
			var freqRes, amountRes, patternRes CheckResult

			// Freq Check
			go func() {
				defer checkWg.Done()
				freqRes = s.frequencyCheck(tx, userTransactionsMap[tx.UserID])
			}()

			// Amount Check
			go func() {
				defer checkWg.Done()
				amountRes = s.amountCheck(tx, userTransactionsMap[tx.UserID])
			}()

			// Pattern Check
			go func() {
				defer checkWg.Done()
				patternRes = s.patternCheck(tx, userTransactionsMap[tx.UserID])
			}()

			checkWg.Wait()

			freqResult := freqRes.Result
			amountResult := amountRes.Result
			patternResult := patternRes.Result

			// Calculate weighted fraud score
			freqScore := freqResult["confidence_score"].(float64)
			amountScore := amountResult["confidence_score"].(float64)
			patternScore := patternResult["confidence_score"].(float64)
			finalScore := (freqScore * 0.4) + (amountScore * 0.3) + (patternScore * 0.3)

			// Determine risk level based on => final score
			var riskLevel string
			switch {
			case finalScore > 80:
				riskLevel = "high"
			case finalScore >= 50:
				riskLevel = "medium"
			default:
				riskLevel = "low"
			}

			detectionResults := map[string]interface{}{
				"frequency_check": freqResult,
				"amount_check":    amountResult,
				"pattern_check":   patternResult,
			}

			processingResult := ProcessingResult{
				FraudResult: FraudResult{
					TransactionID:    tx.OrderID,
					FraudScore:       finalScore,
					RiskLevel:        riskLevel,
					DetectionResults: detectionResults,
				},
				Durations: CheckDurations{
					FreqDuration:    freqRes.Duration,
					AmountDuration:  amountRes.Duration,
					PatternDuration: patternRes.Duration,
				},
			}

			processingResultsChan <- processingResult
		}(tx)
	}

	wg.Wait()
	close(processingResultsChan)

	// Collect all processing results
	var processingResults []ProcessingResult
	for pr := range processingResultsChan {
		processingResults = append(processingResults, pr)
	}

	// Filter results by specified risk levels
	var filteredResults []FraudResult
	for _, pr := range processingResults {
		if len(riskLevels) == 0 || contains(riskLevels, pr.FraudResult.RiskLevel) {
			filteredResults = append(filteredResults, pr.FraudResult)
		}
	}

	// Calculate total and average durations for metadata
	var totalFreqDuration, totalAmountDuration, totalPatternDuration time.Duration
	for _, pr := range processingResults {
		totalFreqDuration += pr.Durations.FreqDuration
		totalAmountDuration += pr.Durations.AmountDuration
		totalPatternDuration += pr.Durations.PatternDuration
	}

	numTransactions := int64(len(processingResults))
	var processingMeta map[string]interface{}

	if numTransactions > 0 {
		avgFreqDuration := totalFreqDuration / time.Duration(numTransactions)
		avgAmountDuration := totalAmountDuration / time.Duration(numTransactions)
		avgPatternDuration := totalPatternDuration / time.Duration(numTransactions)
		totalDuration := time.Since(startTime)

		processingMeta = map[string]interface{}{
			"total_transactions_analyzed": len(transactions),
			"duration_ms":                 totalDuration.Milliseconds(),
			"parallel_tasks": map[string]int64{
				"frequency_analysis_duration_ms": avgFreqDuration.Milliseconds(),
				"amount_analysis_duration_ms":    avgAmountDuration.Milliseconds(),
				"pattern_analysis_duration_ms":   avgPatternDuration.Milliseconds(),
			},
		}
	} else {
		processingMeta = map[string]interface{}{
			"total_transactions_analyzed": 0,
			"duration_ms":                 0,
			"parallel_tasks":              map[string]int64{},
		}
	}

	return BatchResult{
		Transactions:   filteredResults,
		ProcessingMeta: processingMeta,
	}, nil
}

// frequencyCheck analyzes transaction frequency within a one hour time window.
//
// It counts transactions by the same user within one hour of the current transaction
// and assigns a confidence_score based on predefined thresholds.
// Transactions exceeding a frequency threshold are flagged as sus.
func (s *MainServiceImpl) frequencyCheck(tx gen.Transaction, userTxs []gen.Transaction) CheckResult {
	start := time.Now()

	count := int64(0)
	for _, t := range userTxs {
		if t.TransactionDate.Time.Sub(tx.TransactionDate.Time).Abs() <= time.Hour {
			count++
		}
	}

	// Assign confidence score based on transaction count
	var score float64
	switch {
	case count > FrequencyHighThreshold:
		score = 95
	case count > FrequencyMediumHighThreshold:
		score = 85
	case count > FrequencyMediumThreshold:
		score = 75
	case count > FrequencyLowThreshold:
		score = 60
	default:
		score = float64(count) * 10
	}

	isSuspicious := count > FrequencyLowThreshold
	triggers := []string{}

	if isSuspicious {
		triggers = append(triggers, fmt.Sprintf("high order frequency: %d orders in 1 hour", count))
	}

	result := map[string]interface{}{
		"is_suspicious":    isSuspicious,
		"confidence_score": score,
		"triggers":         triggers,
	}

	return CheckResult{Result: result, Duration: time.Since(start)}
}

// amountCheck performs statistical analysis on transaction amounts.
//
// It calculates the mean and standard deviation of historical amounts and computes
// a Z-score for the current transaction. A high Z-score indicates an unusual amount,
// triggering a suspicious flag and confidence score.
func (s *MainServiceImpl) amountCheck(tx gen.Transaction, userTxs []gen.Transaction) CheckResult {
	start := time.Now()

	// collect historical transaction amounts
	var historicalTxs []float64
	for _, t := range userTxs {
		// skip current tx
		if t.ID == tx.ID {
			continue
		}

		amt, err := t.Amount.Float64Value()
		if err != nil {
			continue
		}

		historicalTxs = append(historicalTxs, amt.Float64)
	}

	// Handle insufficient data
	// if not handled, it could lead to NaN which i painfully debug for hours :(
	if len(historicalTxs) < 2 {
		result := map[string]interface{}{
			"is_suspicious":    false,
			"confidence_score": 0.0,
			"triggers":         []string{},
		}

		return CheckResult{Result: result, Duration: time.Since(start)}
	}

	// Calculate mean and standard deviation
	var sum, sumSquared float64

	for _, amt := range historicalTxs {
		sum += amt
		sumSquared += amt * amt
	}

	mean := sum / float64(len(historicalTxs))
	variance := (sumSquared / float64(len(historicalTxs))) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// Get current transaction amount
	currentAmt, err := tx.Amount.Float64Value()
	if err != nil {
		result := map[string]interface{}{
			"is_suspicious":    false,
			"confidence_score": 0.0,
			"triggers":         []string{"invalid amount"},
		}
		return CheckResult{Result: result, Duration: time.Since(start)}
	}

	currentAmount := currentAmt.Float64

	// Handle zero standard deviation
	var score float64
	var isSuspicious bool
	var triggers []string

	// Effectively zero deviation
	if stdDev < 1e-6 {
		if currentAmount > mean {
			score = 100.0
			isSuspicious = true
			triggers = append(triggers, "unusual amount: deviates from uniform historical amounts")
		} else {
			score = 0.0
			isSuspicious = false
		}
	} else {
		zScore := (currentAmount - mean) / stdDev
		score = math.Min(100, math.Max(0, (zScore-2)*20))
		isSuspicious = zScore > 2

		if isSuspicious {
			triggers = append(triggers, fmt.Sprintf("unusual amount: Z-score %.2f", zScore))
		}
	}

	result := map[string]interface{}{
		"is_suspicious":    isSuspicious,
		"confidence_score": score,
		"triggers":         triggers,
	}

	return CheckResult{Result: result, Duration: time.Since(start)}
}

// patternCheck detects significant increases in transaction amounts.
//
// It computes the average of historical amounts and calculates the percentage increase
// of the current transaction. A large increase triggers a suspicious flag and a
// confidence score proportional to the spike.
func (s *MainServiceImpl) patternCheck(tx gen.Transaction, userTxs []gen.Transaction) CheckResult {
	start := time.Now()

	// Collect historical transaction amounts
	var historicalTxs []float64
	for _, t := range userTxs {
		if t.ID == tx.ID {
			continue
		}

		amt, err := t.Amount.Float64Value()
		// Skip if amount is null
		if err != nil {
			continue
		}

		historicalTxs = append(historicalTxs, amt.Float64)
	}

	// Handle insufficient data
	if len(historicalTxs) == 0 {
		result := map[string]interface{}{
			"is_suspicious":    false,
			"confidence_score": 0.0,
			"triggers":         []string{},
		}

		return CheckResult{Result: result, Duration: time.Since(start)}
	}

	// Calculate baseline
	var sum float64
	for _, amt := range historicalTxs {
		sum += amt
	}

	baseline := sum / float64(len(historicalTxs))

	currentAmt, err := tx.Amount.Float64Value()
	if err != nil {
		result := map[string]interface{}{
			"is_suspicious":    false,
			"confidence_score": 0.0,
			"triggers":         []string{"invalid amount"},
		}

		return CheckResult{Result: result, Duration: time.Since(start)}
	}

	currentAmount := currentAmt.Float64

	// Calculate percentage increase
	var percentIncrease float64
	if baseline > 0 { // Avoid division by zero
		percentIncrease = ((currentAmount - baseline) / baseline) * 100
	}

	// Calculate confidence score
	score := math.Min(100, percentIncrease/4)
	isSuspicious := percentIncrease > 300
	triggers := []string{}

	if isSuspicious {
		triggers = append(triggers, fmt.Sprintf("spike in spending: %.0f%% increase", percentIncrease))
	}

	result := map[string]interface{}{
		"is_suspicious":    isSuspicious,
		"confidence_score": score,
		"triggers":         triggers,
	}

	return CheckResult{Result: result, Duration: time.Since(start)}
}

// js inspired contains
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
