package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/study/jobboard/vacancy-service/internal/model"
)

var ErrCompanyNotFound = errors.New("company not found")

// EmployerClient talks to employer-service over HTTP.
type EmployerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewEmployerClient(baseURL string) *EmployerClient {
	return &EmployerClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type employerProfileResponse struct {
	ID          uuid.UUID `json:"id"`
	CompanyName string    `json:"company_name"`
	Description *string   `json:"description"`
	Website     *string   `json:"website"`
	City        *string   `json:"city"`
}

// GetMyEmployerID calls GET /api/v1/employer/profile with the caller's JWT.
func (c *EmployerClient) GetMyEmployerID(ctx context.Context, bearerToken string) (uuid.UUID, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/employer/profile", nil)
	if err != nil {
		return uuid.Nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("employer-service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("employer-service returned %d", resp.StatusCode)
	}

	var body employerProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return uuid.Nil, err
	}
	return body.ID, nil
}

// GetCompany calls public GET /api/v1/employer/companies/{id}.
func (c *EmployerClient) GetCompany(ctx context.Context, employerID uuid.UUID) (model.CompanyInfo, error) {
	url := fmt.Sprintf("%s/api/v1/employer/companies/%s", c.baseURL, employerID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.CompanyInfo{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.CompanyInfo{}, fmt.Errorf("employer-service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return model.CompanyInfo{}, ErrCompanyNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return model.CompanyInfo{}, fmt.Errorf("employer-service returned %d", resp.StatusCode)
	}

	var body employerProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return model.CompanyInfo{}, err
	}
	return model.CompanyInfo{
		ID:          body.ID,
		CompanyName: body.CompanyName,
		Description: body.Description,
		Website:     body.Website,
		City:        body.City,
	}, nil
}
