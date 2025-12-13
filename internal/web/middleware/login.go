package middleware

import (
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePath(paths string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, paths)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if slices.Contains(l.paths, ctx.Request.RequestURI) {
			ctx.Next()
			return
		}

		session := sessions.Default(ctx)
		email := session.Get("userEmail")
		if email == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "not logged in"})
			ctx.Abort()
			return
		}

		session.Options(sessions.Options{
			MaxAge: 3600,
		})

		// Refresh session expiration time
		updateTime := session.Get("update_time")
		now := time.Now().UnixMilli()
		if updateTime == nil {
			log.Println("First refresh session time")
			session.Set("update_time", now)
			session.Save()
			ctx.Next()
			return
		}

		updateTimeValue, ok := updateTime.(int64)
		if !ok {
			log.Println("Session time format error")
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "system error"})
			ctx.Abort()
			return
		}

		if now-updateTimeValue > 60*1000 { // 1 minute no operation, refresh session
			log.Println("Refresh session time")
			session.Set("update_time", now)
			session.Save()
		}

		log.Printf("Current session user email: %v\n", email)
		ctx.Next()
	}
}
