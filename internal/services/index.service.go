package services

import (
	"context"
	"cudo-test/gen"
	"cudo-test/internal/repos"
	"log"
	"sync"
)

type MainService interface {
	GetAllTransactions() ([]gen.Transaction, error)
	DetectFraud() bool
}

type MainServiceImpl struct {
	mainRepo repos.MainRepo
}

func InitMainService(mainRepo repos.MainRepo) *MainServiceImpl {
	return &MainServiceImpl{
		mainRepo: mainRepo,
	}
}

func (s *MainServiceImpl) GetAllTransactions() ([]gen.Transaction, error) {
	return s.mainRepo.GetAll(context.Background())
}

func (s *MainServiceImpl) DetectFraud() bool {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		s.frequencyCheck()
	}()

	wg.Wait()

	return false
}

func (s *MainServiceImpl) frequencyCheck() {
	// get user latest data per 1 hour
	result, err := s.mainRepo.GetUserTransactionWithinTimeframe(context.Background())
	if err != nil {
		log.Println(err.Error())
	}

	log.Println("TEST", result)
}

func amountCheck()  {}
func patternCheck() {}
