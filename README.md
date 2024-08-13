<div align="center">
    <b>Go URL Shortener Service</b>
</div>

## Info

<div>
    <p>A lightweight URL shortener service built using Go, Redis, and Fiber. This service allows you to shorten URLs, customize short URLs, and set expiration times for each shortened link. It also includes rate limiting to prevent abuse.</p>
</div>

<br>

## Running Service 

Clone repository
```shell
git clone https://github.com/0mwa/go-url-shortener.git
cd go-url-shortener
```
Run docker-compose
```shell
docker-compose up -d --build
```

<br>

## API Endpoints

### Shorten URL

Endpoint
```
POST https://localhost:3000/api/v1/
```

Request Body
```JSON
{
  "url": "https://example.com",
  "short": "customShort",   // Optional
  "expiry": 48              // Expiry in hours, Optional
}
```

Response Body
```JSON
{
  "url": "https://example.com",
  "short": "localhost:3000/customShort",
  "expiry": 24,             // In hours
  "rate_limit": 10,
  "rate_limit_rest": 30     // In minutes
}
```

---

### Resolve URL

Endpoint
```
GET https://localhost:3000/:url
```



