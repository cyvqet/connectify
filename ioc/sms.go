package ioc

import "connectify/internal/service/sms"

func InitSmsService() sms.Service {
	// use mock sms service
	return sms.NewMockService()
}
