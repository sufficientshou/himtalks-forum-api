package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type SpotifyService struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	ExpiresAt    time.Time
	mu           sync.Mutex
}

type SpotifyToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SpotifyErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type SpotifySearchResponse struct {
	Tracks struct {
		Items []SpotifyTrack `json:"items"`
	} `json:"tracks"`
}

type SpotifyTrack struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URI        string `json:"uri"`
	PreviewURL string `json:"preview_url"`
	DurationMS int    `json:"duration_ms"`
	Artists    []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		Name   string `json:"name"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	} `json:"album"`
}

func NewSpotifyService(clientID, clientSecret string) *SpotifyService {
	return &SpotifyService{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

func (s *SpotifyService) GetToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.AccessToken != "" && time.Now().Before(s.ExpiresAt) {
		return s.AccessToken, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(s.ClientID + ":" + s.ClientSecret))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request Spotify token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response body: %w", err)
	}

	// Check HTTP status before parsing JSON
	if resp.StatusCode != http.StatusOK {
		// Try to parse Spotify error response
		var spotifyErr SpotifyErrorResponse
		if jsonErr := json.Unmarshal(body, &spotifyErr); jsonErr == nil && spotifyErr.Error != "" {
			log.Printf("[Spotify] Token request failed (HTTP %d): %s - %s", resp.StatusCode, spotifyErr.Error, spotifyErr.ErrorDescription)
			return "", fmt.Errorf("Spotify token error (HTTP %d): %s - %s", resp.StatusCode, spotifyErr.Error, spotifyErr.ErrorDescription)
		}
		// If not JSON, include raw body snippet
		bodySnippet := string(body)
		if len(bodySnippet) > 200 {
			bodySnippet = bodySnippet[:200] + "..."
		}
		log.Printf("[Spotify] Token request failed (HTTP %d): %s", resp.StatusCode, bodySnippet)
		return "", fmt.Errorf("Spotify token request failed (HTTP %d): %s", resp.StatusCode, bodySnippet)
	}

	var token SpotifyToken
	err = json.Unmarshal(body, &token)
	if err != nil {
		log.Printf("[Spotify] Failed to parse token response: %v, body: %s", err, string(body))
		return "", fmt.Errorf("failed to parse Spotify token response: %w", err)
	}

	if token.AccessToken == "" {
		log.Printf("[Spotify] Token response has empty access_token, body: %s", string(body))
		return "", fmt.Errorf("Spotify returned empty access token")
	}

	s.AccessToken = token.AccessToken
	s.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn-60) * time.Second)

	log.Printf("[Spotify] Successfully obtained access token (expires in %ds)", token.ExpiresIn)
	return s.AccessToken, nil
}

func (s *SpotifyService) SearchTracks(query string, limit int) ([]SpotifyTrack, error) {
	token, err := s.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get Spotify token: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	searchURL := fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=%d",
		url.QueryEscape(query), limit)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Spotify search API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response body: %w", err)
	}

	// Check HTTP status before parsing JSON
	if resp.StatusCode != http.StatusOK {
		// If token expired (401), clear cached token so next request gets a fresh one
		if resp.StatusCode == http.StatusUnauthorized {
			s.mu.Lock()
			s.AccessToken = ""
			s.ExpiresAt = time.Time{}
			s.mu.Unlock()
			log.Printf("[Spotify] Search returned 401, cleared cached token. Will retry on next request.")
		}
		bodySnippet := string(body)
		if len(bodySnippet) > 200 {
			bodySnippet = bodySnippet[:200] + "..."
		}
		log.Printf("[Spotify] Search request failed (HTTP %d): %s", resp.StatusCode, bodySnippet)
		return nil, fmt.Errorf("Spotify search failed (HTTP %d): %s", resp.StatusCode, bodySnippet)
	}

	var searchResponse SpotifySearchResponse
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		log.Printf("[Spotify] Failed to parse search response: %v, body: %s", err, string(body))
		return nil, fmt.Errorf("failed to parse Spotify search response: %w", err)
	}

	return searchResponse.Tracks.Items, nil
}

func (s *SpotifyService) GetTrack(trackID string) (*SpotifyTrack, error) {
	token, err := s.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get Spotify token: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	trackURL := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackID)

	req, err := http.NewRequest("GET", trackURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create track request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Spotify track API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read track response body: %w", err)
	}

	// Check HTTP status before parsing JSON
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			s.mu.Lock()
			s.AccessToken = ""
			s.ExpiresAt = time.Time{}
			s.mu.Unlock()
		}
		bodySnippet := string(body)
		if len(bodySnippet) > 200 {
			bodySnippet = bodySnippet[:200] + "..."
		}
		log.Printf("[Spotify] GetTrack request failed (HTTP %d): %s", resp.StatusCode, bodySnippet)
		return nil, fmt.Errorf("Spotify get track failed (HTTP %d): %s", resp.StatusCode, bodySnippet)
	}

	var track SpotifyTrack
	err = json.Unmarshal(body, &track)
	if err != nil {
		log.Printf("[Spotify] Failed to parse track response: %v, body: %s", err, string(body))
		return nil, fmt.Errorf("failed to parse Spotify track response: %w", err)
	}

	return &track, nil
}
