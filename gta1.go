package vision

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// coordinateResponse is the JSON response from the GTA1 service.
type coordinateResponse struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// httpClient abstracts HTTP requests for testing.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GTA1Resolver communicates with the GTA1 vision service.
type GTA1Resolver struct {
	baseURL string
	client  httpClient
}

// NewGTA1Resolver creates a new GTA1 resolver.
func NewGTA1Resolver(baseURL string) *GTA1Resolver {
	return &GTA1Resolver{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Resolve sends a screenshot and locate instruction to the GTA1 service
// and returns the detected coordinates.
func (g *GTA1Resolver) Resolve(ctx context.Context, locate string, screenshot []byte) (Coordinates, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("instruction", locate); err != nil {
		return Coordinates{}, fmt.Errorf("write instruction field: %w", err)
	}

	part, err := writer.CreateFormFile("image_file", "screenshot.png")
	if err != nil {
		return Coordinates{}, fmt.Errorf("create image field: %w", err)
	}
	if _, err := part.Write(screenshot); err != nil {
		return Coordinates{}, fmt.Errorf("write image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return Coordinates{}, fmt.Errorf("close multipart writer: %w", err)
	}

	url := g.baseURL + "/process/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return Coordinates{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.client.Do(req)
	if err != nil {
		return Coordinates{}, fmt.Errorf("%w: %v", ErrServiceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return Coordinates{}, fmt.Errorf("%w: status %d: %s", ErrServiceUnavailable, resp.StatusCode, string(respBody))
	}

	var result coordinateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Coordinates{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return Coordinates{
		X: int(result.X),
		Y: int(result.Y),
	}, nil
}
