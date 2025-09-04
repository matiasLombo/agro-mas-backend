package products

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type LocationService struct {
	client *http.Client
	apiURL string
}

type Province struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

type Department struct {
	ID        string   `json:"id"`
	Nombre    string   `json:"nombre"`
	Provincia Province `json:"provincia"`
}

type Settlement struct {
	ID           string     `json:"id"`
	Nombre       string     `json:"nombre"`
	Provincia    Province   `json:"provincia"`
	Departamento Department `json:"departamento"`
}

type ProvincesResponse struct {
	Provincias []Province `json:"provincias"`
}

type DepartmentsResponse struct {
	Departamentos []Department `json:"departamentos"`
}

type SettlementsResponse struct {
	Asentamientos []Settlement `json:"asentamientos"`
}

func NewLocationService() *LocationService {
	return &LocationService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiURL: "https://apis.datos.gob.ar/georef/api",
	}
}

// GetProvinceByID gets province name by ID
func (ls *LocationService) GetProvinceByID(ctx context.Context, provinceID string) (string, error) {
	if provinceID == "" {
		return "", nil
	}

	url := fmt.Sprintf("%s/provincias?id=%s", ls.apiURL, url.QueryEscape(provinceID))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ls.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response ProvincesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Provincias) > 0 {
		return response.Provincias[0].Nombre, nil
	}

	return "", fmt.Errorf("province not found")
}

// GetSettlementByName gets settlement info by name and province
func (ls *LocationService) GetSettlementByName(ctx context.Context, settlementName, provinceID string) (*Settlement, error) {
	if settlementName == "" || provinceID == "" {
		return nil, nil
	}

	// First get the province name
	provinceName, err := ls.GetProvinceByID(ctx, provinceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get province name: %w", err)
	}

	url := fmt.Sprintf("%s/asentamientos?nombre=%s&provincia=%s&max=1",
		ls.apiURL,
		url.QueryEscape(settlementName),
		url.QueryEscape(provinceName))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ls.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response SettlementsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Asentamientos) > 0 {
		return &response.Asentamientos[0], nil
	}

	return nil, fmt.Errorf("settlement not found")
}

// ResolveLocationNames resolves province and settlement codes/names to full location info
func (ls *LocationService) ResolveLocationNames(ctx context.Context, provinceID, cityName string) (provinceName, departmentName, settlementName string, err error) {
	// Get province name
	if provinceID != "" {
		provinceName, err = ls.GetProvinceByID(ctx, provinceID)
		if err != nil {
			// Log error but don't fail completely
			fmt.Printf("Warning: failed to get province name for ID %s: %v\n", provinceID, err)
		}
	}

	// Get settlement info (includes department)
	if cityName != "" && provinceID != "" {
		settlement, err := ls.GetSettlementByName(ctx, cityName, provinceID)
		if err != nil {
			// Log error but don't fail completely
			fmt.Printf("Warning: failed to get settlement info for %s in province %s: %v\n", cityName, provinceID, err)
		} else if settlement != nil {
			// Use the settlement name from API (normalized)
			settlementName = settlement.Nombre
			departmentName = settlement.Departamento.Nombre
			// Override province name with the one from settlement (more accurate)
			if settlement.Provincia.Nombre != "" {
				provinceName = settlement.Provincia.Nombre
			}
		}
	}

	return provinceName, departmentName, settlementName, nil
}
