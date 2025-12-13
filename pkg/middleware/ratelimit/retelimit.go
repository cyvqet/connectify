package ratelimit

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Builder is a rate limiter builder, used to create a rate limiting middleware based on the sliding window algorithm
type Builder struct {
	prefix   string        // Redis key prefix
	cmd      redis.Cmdable // Redis client
	interval time.Duration // Time window length
	rate     int           // Maximum number of requests allowed within the window
}

//go:embed slide_window.lua
var luaScript string // Embedded sliding window Lua script

// NewBuilder creates a Builder instance
func NewBuilder(cmd redis.Cmdable, interval time.Duration, rate int) *Builder {
	return &Builder{
		cmd:      cmd,
		prefix:   "ip-limiter", // Default prefix
		interval: interval,
		rate:     rate,
	}
}

// Prefix sets the Redis key prefix
func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

// Build creates a Gin rate limiting middleware
func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			log.Println(err)
			// Conservative approach (rate limiting) vs aggressive approach (allowing through)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

// limit performs rate limiting checks
func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	// Construct Redis key based on client IP
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	// Execute Lua script to implement atomic sliding window rate limiting
	return b.cmd.Eval(ctx, luaScript, []string{key},
		b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}
