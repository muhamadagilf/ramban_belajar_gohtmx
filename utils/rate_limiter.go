// Package utils
package utils

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
)

type TokenContainer struct {
	Tokens     float64
	LastRefill int64
}

type LimiterConfig struct {
	RateLimit, TokenCapacity float64
}

var (
	rateLimiterConfig = map[string]LimiterConfig{
		"POST /login": {
			RateLimit:     5.0 / 60.0,
			TokenCapacity: 5.0,
		},

		"POST /admin/login": {
			RateLimit:     3.0 / 60.0,
			TokenCapacity: 3.0,
		},

		"userpublic": {
			RateLimit:     100.0 / 60.0,
			TokenCapacity: 100.0,
		},

		"useradmin": {
			RateLimit:     500.0 / 60.0,
			TokenCapacity: 500.0,
		},
	}

	userTokenContainers = make(map[uuid.UUID]*TokenContainer)
	apiTokenContainers  = make(map[string]*TokenContainer)
	mu                  = &sync.Mutex{}
)

func (container *TokenContainer) refillTokens(now int64, rate, tokenCapacity float64) {
	timePassed := float64(now-container.LastRefill) / 1000.0
	container.Tokens = math.Min(tokenCapacity, container.Tokens+(timePassed*rate))
	container.LastRefill = now
}

func MiddlewareUserRateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("claims").(*server.Claims)
		if !ok {
			return next(c)
		}

		var userLimiterConfig LimiterConfig
		currentTime := time.Now().UnixMilli()

		switch claims.Roles[0] {
		case USER_ROLE_ADMIN:
			userLimiterConfig = rateLimiterConfig["useradmin"]
		default:
			userLimiterConfig = rateLimiterConfig["userpublic"]
		}

		// LOCK THE OPPERATION, SO IT DOES NOT HAVE RACE CONDITION ISSUE
		mu.Lock()
		defer mu.Unlock()

		userTokenContainer, ok := userTokenContainers[claims.UserID]
		if !ok {
			userTokenContainer = &TokenContainer{
				Tokens:     userLimiterConfig.TokenCapacity,
				LastRefill: currentTime,
			}

			userTokenContainers[claims.UserID] = userTokenContainer
		}

		// REFILL THE TOKENS, BASED ON HOW LONG THE TIMESPAN FROM LASTREFIIL
		userTokenContainer.refillTokens(
			currentTime,
			userLimiterConfig.RateLimit,
			userLimiterConfig.TokenCapacity,
		)

		// IF USER'S TOKEN RUNOUT, SEND "StatusTooManyRequests"
		// HOWEVER, IF USER'S TOKEN STILL VALID, TAKE 1.0 PER-REQUEST
		if userTokenContainer.Tokens < 1.0 {
			return c.String(
				http.StatusTooManyRequests,
				"Rate Limit Exceeded, please try another minute",
			)
		}

		userTokenContainer.Tokens -= 1.0
		return next(c)
	}
}

func MiddlewareAPIRateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		resourceKey := fmt.Sprintf("%s %s", c.Request().Method, c.Path())
		currentTime := time.Now().UnixMilli()
		userIP := c.RealIP()
		if userIP == "" {
			log.Println("Internal Server Error, at code:41500")
			return c.String(
				http.StatusInternalServerError,
				"Internal Server Error, Contact Suport with code:41500",
			)
		}

		apiTokenLimiterKey := userIP + ":" + resourceKey
		log.Println(apiTokenLimiterKey)

		if apiLimiterConfig, ok := rateLimiterConfig[resourceKey]; ok {

			mu.Lock()
			defer mu.Unlock()

			apiTokenContainer, ok := apiTokenContainers[apiTokenLimiterKey]
			if !ok {
				apiTokenContainer = &TokenContainer{
					Tokens:     apiLimiterConfig.TokenCapacity,
					LastRefill: currentTime,
				}

				apiTokenContainers[apiTokenLimiterKey] = apiTokenContainer
			}

			log.Println(apiTokenLimiterKey, apiTokenContainer)

			apiTokenContainer.refillTokens(
				currentTime,
				apiLimiterConfig.RateLimit,
				apiLimiterConfig.TokenCapacity,
			)

			log.Println(apiTokenLimiterKey, apiTokenContainer)

			if apiTokenContainer.Tokens < 1.0 {
				return c.String(
					http.StatusTooManyRequests,
					"Rate Limit Exceeded, Please Try another minute",
				)
			}

			apiTokenContainer.Tokens -= 1.0
			log.Println(apiTokenLimiterKey, apiTokenContainer)
		}

		return next(c)
	}
}

func CleanupLimiterContainersWatcher() {
	go func() {
		log.Println("CLEANER RUNNING: Stale Limiter Token Containers")
		for {
			// GET CURRENT_TIME
			now := time.Now()

			// GET THE TIME SCHEDULE TO CLEANUP (00:00)
			// THE SCHEDULE SHOULD CALCULATE BASED ON CURRENT_TIME (NOW)
			// SO, IF NOW (2025-11-14 13:00) THE SCHEDULE SHOULD (2025-11-15 00:00)
			// IT SHOULD THE NEXT DAY,
			// WHICH MEAN EVERY ITERATION WOULD UPDATE THE NOW & THE CLEANUP SCHEDULE
			// AGAIN, BASED ON THE "NOW"" TIME
			cleanupSchedule := time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				0, 0, 0, 0,
				now.Location(),
			).Add(24 * time.Hour)

			// THEN, GET THE DURATION UNTIL SCHEDULE TIME
			duration := time.Until(cleanupSchedule)

			// BLOCK THE FLOW, UNTIL THE DURATION ELAPSED
			// THEN "time.After" WOULD SEND THE CURRENT_TIME TO "<-chan time.Time"
			<-time.After(duration)
			log.Println("CLEANER CHECKPOINT: Limiter Token Containers")

			// ONCE, THE "time.Time" SEND THE "chan"
			// CLEANS UP GET TO WORKING
			// GET ONE HOUR BEFORE CLEANUP SCHEDULE
			// TO CHECK, ONLY DELETE THE TOKEN LIMITER, THAT HASNT BEEN REFILL WITHIN ONE HOUR BEFORE SCHEDULE
			oneHourBeforeSchedule := time.Now().Add(-1 * time.Hour).UnixMilli()
			for key, value := range userTokenContainers {
				if value.LastRefill < oneHourBeforeSchedule {
					delete(userTokenContainers, key)
				}
			}

			for key, value := range apiTokenContainers {
				if value.LastRefill < oneHourBeforeSchedule {
					delete(apiTokenContainers, key)
				}
			}

			log.Println("LIMITER TOKEN CONTAINERS, CLEANED UP")

		}
	}()
}
