package routes

import (
	"errors"
	"github.com/0mwa/go-url-shortener/database"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {
	shortURL := c.Params("url")

	rDatabase := database.NewRedisClient(0)
	defer rDatabase.Close()

	rRequestLimitCounter := database.NewRedisClient(1)
	defer rRequestLimitCounter.Close()

	originalURL, err := rDatabase.Get(database.Ctx, shortURL).Result()
	if errors.Is(err, redis.Nil) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short URL not found in the database"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to database"})
	}

	if err = rRequestLimitCounter.Incr(database.Ctx, "counter").Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to increment counter"})
	}

	return c.Redirect(originalURL, fiber.StatusMovedPermanently)
}
