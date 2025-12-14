package sms

import (
	"context"

	"go.uber.org/zap"
)

//go:generate mockgen -source=sms.go -destination=mocks/sms_mock.go -package=smsmocks

type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}

type MockService struct{}

func NewMockService() Service {
	return &MockService{}
}

func (s *MockService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	zap.L().Info("mock send sms",
		zap.String("tplId", tplId),
		zap.Strings("args", args),
		zap.Strings("numbers", numbers),
	)
	return nil
}
