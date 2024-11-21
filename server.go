package main

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type URLEntry struct {
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	VisitCount  int       `json:"visit_count"`
}

type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
}

type ShortenRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type URLShortener struct {
	urls map[string]URLEntry
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]URLEntry),
	}
}

func (s *URLShortener) GenerateShortCode(url string) string {
	hash := sha256.Sum256([]byte(url + time.Now().String()))

	shortCode := base64.URLEncoding.EncodeToString(hash[:])[:8]
	return shortCode
}

func (s *URLShortener) ShortenURL(c echo.Context) error {
	var req ShortenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid URL format")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	shortcode := s.GenerateShortCode(req.URL)
	entry := URLEntry{
		OriginalURL: req.URL,
		ShortCode:   shortcode,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		VisitCount:  0,
	}
	s.urls[shortcode] = entry
	baseURL := c.Scheme() + "://" + c.Request().Host
	respone := URLResponse{
		ShortURL:    baseURL + "/r/" + shortcode,
		OriginalURL: req.URL,
		ShortCode:   shortcode,
	}
	return c.JSON(http.StatusCreated, respone)
}

func (s *URLShortener) handleRedirect(c echo.Context) error {
	shortCode := c.Param("code")
	s.mu.Lock()
	entry, exists := s.urls[shortCode]
	if exists {
		entry.VisitCount++
		entry.LastUsedAt = time.Now()
		s.urls[shortCode] = entry
	}
	s.mu.Unlock()
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "URL not found")
	}
	return c.Redirect(http.StatusTemporaryRedirect, entry.OriginalURL)
}

func (s *URLShortener) handleListURLs(c echo.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urls := make([]URLEntry, 0, len(s.urls))
	for _, entry := range s.urls {
		urls = append(urls, entry)
	}
	return c.JSON(http.StatusOK, urls)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {

	//create Echo instance

	e := echo.New()

	e.HTTPErrorHandler = errorHandler

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	//set up validator
	e.Validator = &CustomValidator{validator: validator.New()}
	shortener := NewURLShortener()

	api := e.Group("api")
	{
		api.POST("/shorten", shortener.ShortenURL)
		api.GET("/urls", shortener.handleListURLs)
	}
	e.GET("/r/:code", shortener.handleRedirect)
	e.Logger.Fatal(e.Start(":8080"))

}

func errorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message.(string)
	}
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]string{"error": msg})
		}
		if err != nil {
			c.Logger().Error(err)
		}
	}
}
