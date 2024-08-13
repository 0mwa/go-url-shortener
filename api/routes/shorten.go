package routes

import (
	"errors"
	"fmt"
	"github.com/0mwa/go-url-shortener/database"
	"github.com/0mwa/go-url-shortener/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"os"
	"strconv"
	"time"
)

type Request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type Response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_rest"`
}

func ShortenURL(c *fiber.Ctx) error {
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	rRequestLimitCounter := database.NewRedisClient(1)
	defer rRequestLimitCounter.Close()

	rDatabase := database.NewRedisClient(0)
	defer rDatabase.Close()

	if limit, err := handleRateLimiting(rRequestLimitCounter, c.IP()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": err.Error(), "rate_limit": limit})
	}

	if status, err := validateURL(&req); err != nil {
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}

	id := generateShortId(req.CustomShort)

	if err := saveURLInRedis(rDatabase, id, req.URL, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response, err := createResponse(rRequestLimitCounter, req, id, c.IP())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create response"})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func handleRateLimiting(rdb *redis.Client, ip string) (time.Duration, error) {
	value, err := rdb.Get(database.Ctx, ip).Result()
	if errors.Is(err, redis.Nil) {
		_ = rdb.Set(database.Ctx, ip, os.Getenv("API_QUOTA"), 30*time.Minute).Err()
	} else {
		valueInt, _ := strconv.Atoi(value)
		if valueInt <= 0 {
			limit, _ := getRateLimitReset(rdb, ip)
			return limit, fmt.Errorf("rate limit exceeded")
		}
	}
	return 0, nil
}

func validateURL(req *Request) (int, error) {
	if !govalidator.IsURL(req.URL) {
		return fiber.StatusBadRequest, fmt.Errorf("invalid URL")
	}

	if !helpers.DomainError(req.URL) {
		return fiber.StatusInternalServerError, fmt.Errorf("domain error")
	}

	req.URL = helpers.EnforceHTTP(req.URL)
	return fiber.StatusOK, nil
}

func generateShortId(customShort string) string {
	if customShort != "" {
		return customShort
	}
	return uuid.New().String()[:6]
}

func saveURLInRedis(rdb *redis.Client, id string, url string, req *Request) error {
	if req.Expiry == 0 {
		req.Expiry = 24 * time.Hour
	}

	err := rdb.Set(database.Ctx, id, url, req.Expiry).Err()
	if err != nil {
		return fmt.Errorf("unable to connect to server")
	}

	return nil
}

func createResponse(rdb *redis.Client, req Request, id, ip string) (Response, error) {
	domain := os.Getenv("DOMAIN")
	customShort := domain + "/" + id

	rdb.Decr(database.Ctx, ip)

	remainingRequests, err := getRemainingRequests(rdb, ip)
	if err != nil {
		return Response{}, err
	}

	rateLimitRest, err := getRateLimitReset(rdb, ip)
	if err != nil {
		return Response{}, err
	}

	return Response{
		URL:             req.URL,
		CustomShort:     customShort,
		Expiry:          req.Expiry / time.Hour,
		XRateRemaining:  remainingRequests,
		XRateLimitReset: rateLimitRest,
	}, nil
}

func getRemainingRequests(rdb *redis.Client, ip string) (int, error) {
	value, err := rdb.Get(database.Ctx, ip).Result()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func getRateLimitReset(rdb *redis.Client, ip string) (time.Duration, error) {
	ttl, err := rdb.TTL(database.Ctx, ip).Result()
	if err != nil {
		return 0, err
	}
	return ttl / time.Minute / time.Nanosecond, nil
}
