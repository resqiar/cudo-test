package services

import (
	"context"
	"cudo-test/internal/repos"
	"log"
	"math"
	"sync"
)

type MainService interface {
	DetectFraud(userID int64) *FraudResult
}

type MainServiceImpl struct {
	mainRepo repos.MainRepo
}

type FraudResult struct {
	FrequencyScore float64
	AmountScore    float64
	PatternScore   float64
	FinalScore     float64
	RiskLevel      string
}

func InitMainService(mainRepo repos.MainRepo) *MainServiceImpl {
	return &MainServiceImpl{
		mainRepo: mainRepo,
	}
}

// DetectFraud performs fraud detection by running multiple risk assessment checks
// (Frequency, Amount, and Pattern checks) in parallel. It then aggregates the results
// to compute a final fraud score and determine the user's risk level.
//
// 1. Runs each check in a separate goroutine for performance.
// 2. Aggregates the results once all checks are completed.
// 3. Computes a weighted final score based on the results.
// 4. Assigns a risk level based on the final fraud score.
func (s *MainServiceImpl) DetectFraud(userID int64) *FraudResult {
	var wg sync.WaitGroup

	resultChan := make(chan FraudResult, 3)
	ctx := context.Background()
	wg.Add(3)

	// frequency check
	go func() {
		defer wg.Done()
		freqScore := s.frequencyCheck(ctx, userID)
		resultChan <- FraudResult{FrequencyScore: freqScore}
	}()

	// amount check
	go func() {
		defer wg.Done()
		amountScore := s.amountCheck(ctx, userID)
		resultChan <- FraudResult{AmountScore: amountScore}
	}()

	// pattern check
	go func() {
		defer wg.Done()
		patternScore := s.patternCheck(ctx, userID)
		resultChan <- FraudResult{PatternScore: patternScore}
	}()

	wg.Wait()
	close(resultChan)

	finalResult := FraudResult{}
	for res := range resultChan {
		finalResult.FrequencyScore += res.FrequencyScore
		finalResult.AmountScore += res.AmountScore
		finalResult.PatternScore += res.PatternScore
	}

	// Compute the final fraud score using weighted contributions:
	// - Frequency Score: 40% weight
	// - Amount Score: 30% weight
	// - Pattern Score: 30% weight
	finalResult.FinalScore = (finalResult.FrequencyScore * 0.4) +
		(finalResult.AmountScore * 0.3) +
		(finalResult.PatternScore * 0.3)

	// Assign a risk level based on the final score:
	// - High risk: Score > 80
	// - Medium risk: Score between 50 and 80
	// - Low risk: Score < 50
	switch {
	case finalResult.FinalScore > 80:
		finalResult.RiskLevel = "High"
	case finalResult.FinalScore >= 50:
		finalResult.RiskLevel = "Medium"
	default:
		finalResult.RiskLevel = "Low"
	}

	return &finalResult
}

// calculates a risk score based on the number of transactions
// a user has made within a specific timeframe.
//
// - If transactions exceed 8, return 95 (high risk).
// - If transactions exceed 7, return 85.
// - If transactions exceed 6, return 75.
// - If transactions exceed 5, return 60.
// - Otherwise, return transaction count * 10 as a percentage.
func (s *MainServiceImpl) frequencyCheck(ctx context.Context, userID int64) float64 {
	transactions, err := s.mainRepo.GetUserTransactionWithinTimeframe(ctx, userID)
	if err != nil {
		log.Printf("Frequency check error: %v", err)
		return 0
	}

	if len(transactions) == 0 {
		return 0
	}

	transacCount := transactions[0].TransacCount

	switch {
	case transacCount > 8:
		return 95
	case transacCount > 7:
		return 85
	case transacCount > 6:
		return 75
	case transacCount > 5:
		return 60
	default:
		return float64(transacCount) * 10
	}
}

// calculates a risk score based on the transaction amount pattern.
//
// 1. Compute the mean and standard deviation of all transactions.
// 2. Calculate the Z-score for the latest transaction to detect anomalies.
// 3. Normalize the score between 0-100 for risk assessment.
func (s *MainServiceImpl) amountCheck(ctx context.Context, userID int64) float64 {
	txs, err := s.mainRepo.GetUserTransactions(ctx, userID)
	if err != nil {
		log.Printf("Amount check error: %v", err)
		return 0
	}

	if len(txs) == 0 {
		return 0
	}

	var sum, sumSquared float64

	// calculate sum and squared sum
	for _, tx := range txs {
		// this is type cast from SQLC,
		// a bit pain but have to be done
		amt, _ := tx.Amount.Float64Value()
		sum += amt.Float64
		sumSquared += amt.Float64 * amt.Float64
	}

	// determine mean and std
	n := float64(len(txs))
	mean := sum / n
	variance := (sumSquared / n) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// prevent division by 0
	if stdDev == 0 {
		return 0
	}

	// this is type cast from SQLC,
	// a bit pain but have to be done
	latestAmount, _ := txs[0].Amount.Float64Value()

	// z-score to determine how far latest transaction deviates from the mean
	zScore := (latestAmount.Float64 - mean) / stdDev

	// normalize score between 0-100
	// higher deviation means higher risk
	return math.Min(100, math.Max(0, (zScore-2)*20))
}

// detects unusual transaction behavior by comparing the latest transaction amount
// against historical averages.
//
// 1. Calculate the baseline average amount from past transactions (excluding the latest).
// 2. Compare the latest transaction to this baseline and determine the percentage increase.
// 3. Normalize the result to a 0-100 risk score.
func (s *MainServiceImpl) patternCheck(ctx context.Context, userID int64) float64 {
	txs, err := s.mainRepo.GetUserTransactions(ctx, userID)
	if err != nil {
		log.Printf("Pattern check error: %v", err)
		return 0
	}

	// need at least 2 transactions to compare patterns
	if len(txs) < 2 {
		return 0
	}

	// compute baseline average from all past transactions (excluding latest)
	var baselineSum float64
	for i := 1; i < len(txs); i++ {
		amt, _ := txs[i].Amount.Float64Value()
		baselineSum += amt.Float64
	}
	baseline := baselineSum / float64(len(txs)-1)

	// get latest transaction amount
	latestAmount, _ := txs[0].Amount.Float64Value()

	// calculate percentage increase from baseline
	percentIncrease := ((latestAmount.Float64 - baseline) / baseline) * 100

	score := percentIncrease / 4

	return math.Max(0, math.Min(100, score))
}
