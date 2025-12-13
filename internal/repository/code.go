package repository

import (
	"context"

	"connectify/internal/repository/cache"
)

var (
	ErrVerificationCodeSendRateLimited  = cache.ErrVerificationCodeSendRateLimited
	ErrVerificationCodeCheckRateLimited = cache.ErrVerificationCodeCheckRateLimited
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository(codeCache *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cache: codeCache,
	}
}

// Set saves the verification code for the specified business type and phone number.
func (r *CodeRepository) Set(ctx context.Context, bizType, phone, verificationCode string) error {
	return r.cache.Set(ctx, bizType, phone, verificationCode)
}

// Verify verifies the verification code is correct
func (r *CodeRepository) Verify(ctx context.Context, bizType, phone, verificationCode string) (bool, error) {
	return r.cache.Verify(ctx, bizType, phone, verificationCode)
}
