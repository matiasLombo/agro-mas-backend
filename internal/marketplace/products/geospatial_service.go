package products

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"github.com/gin-gonic/gin"
)

type GeospatialService struct {
	db *sql.DB
}

type NearbySearchRequest struct {
	Latitude    float64     `json:"latitude" binding:"required"`
	Longitude   float64     `json:"longitude" binding:"required"`
	RadiusKm    float64     `json:"radius_km" binding:"required,min=1,max=500"`
	Category    string      `json:"category,omitempty"`
	Subcategory string      `json:"subcategory,omitempty"`
	MaxResults  int         `json:"max_results,omitempty"`
	PriceRange  *PriceRange `json:"price_range,omitempty"`
}

type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type NearbyProduct struct {
	Product    *Product `json:"product"`
	DistanceKm float64  `json:"distance_km"`
	BearingDeg float64  `json:"bearing_deg,omitempty"`
}

type LocationBounds struct {
	NorthEast Point `json:"north_east"`
	SouthWest Point `json:"south_west"`
}

type GeospatialStats struct {
	TotalProductsInRadius int            `json:"total_products_in_radius"`
	ProductsByCategory    map[string]int `json:"products_by_category"`
	ProductsByProvince    map[string]int `json:"products_by_province"`
	AverageDistance       float64        `json:"average_distance_km"`
	FarthestProduct       *NearbyProduct `json:"farthest_product,omitempty"`
	ClosestProduct        *NearbyProduct `json:"closest_product,omitempty"`
}

func NewGeospatialService(db *sql.DB) *GeospatialService {
	return &GeospatialService{db: db}
}

// FindNearbyProducts finds products within a specified radius
func (g *GeospatialService) FindNearbyProducts(ctx context.Context, req *NearbySearchRequest) ([]*NearbyProduct, error) {
	// Build the SQL query with PostGIS functions
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.description, p.category, p.subcategory,
			p.price, p.price_type, p.currency, p.unit, p.quantity,
			p.province, p.city, p.seller_name, p.seller_phone, p.seller_rating,
			p.seller_verification_level, p.views_count, p.favorites_count,
			p.created_at, p.published_at,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[0] ELSE NULL END as lng,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[1] ELSE NULL END as lat,
			ST_Distance(
				ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
				ST_GeogFromText(ST_AsText(p.location_coordinates))
			) / 1000 as distance_km,
			ST_Azimuth(
				ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
				ST_GeogFromText(ST_AsText(p.location_coordinates))
			) * 180 / PI() as bearing_deg
		FROM products p
		WHERE p.is_active = true 
		AND p.location_coordinates IS NOT NULL
		AND ST_DWithin(
			ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
			ST_GeogFromText(ST_AsText(p.location_coordinates)),
			$3
		)`

	args := []interface{}{req.Longitude, req.Latitude, req.RadiusKm * 1000} // Convert km to meters
	argIndex := 4

	// Add category filter
	if req.Category != "" {
		query += fmt.Sprintf(" AND p.category = $%d", argIndex)
		args = append(args, req.Category)
		argIndex++
	}

	// Add subcategory filter
	if req.Subcategory != "" {
		query += fmt.Sprintf(" AND p.subcategory = $%d", argIndex)
		args = append(args, req.Subcategory)
		argIndex++
	}

	// Add price range filter
	if req.PriceRange != nil {
		if req.PriceRange.Min > 0 {
			query += fmt.Sprintf(" AND p.price >= $%d", argIndex)
			args = append(args, req.PriceRange.Min)
			argIndex++
		}
		if req.PriceRange.Max > 0 {
			query += fmt.Sprintf(" AND p.price <= $%d", argIndex)
			args = append(args, req.PriceRange.Max)
			argIndex++
		}
	}

	// Order by distance
	query += " ORDER BY distance_km ASC"

	// Limit results
	maxResults := req.MaxResults
	if maxResults <= 0 || maxResults > 100 {
		maxResults = 50 // Default limit
	}
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, maxResults)

	rows, err := g.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute nearby search query: %w", err)
	}
	defer rows.Close()

	var nearbyProducts []*NearbyProduct
	for rows.Next() {
		product := &Product{}
		var lng, lat, distanceKm, bearingDeg sql.NullFloat64

		err := rows.Scan(
			&product.ID, &product.UserID, &product.Title, &product.Description,
			&product.Category, &product.Subcategory, &product.Price, &product.PriceType,
			&product.Currency, &product.Unit, &product.Quantity, &product.Province,
			&product.City, &product.SellerName, &product.SellerPhone, &product.SellerRating,
			&product.SellerVerificationLevel, &product.ViewsCount, &product.FavoritesCount,
			&product.CreatedAt, &product.PublishedAt, &lng, &lat, &distanceKm, &bearingDeg)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}

		// Set coordinates
		if lng.Valid && lat.Valid {
			product.LocationCoordinates = &Point{
				Lng: lng.Float64,
				Lat: lat.Float64,
			}
		}

		nearbyProduct := &NearbyProduct{
			Product:    product,
			DistanceKm: distanceKm.Float64,
		}

		if bearingDeg.Valid {
			nearbyProduct.BearingDeg = bearingDeg.Float64
		}

		nearbyProducts = append(nearbyProducts, nearbyProduct)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate nearby products: %w", err)
	}

	return nearbyProducts, nil
}

// GetGeospatialStats returns statistics about products in a geographic area
func (g *GeospatialService) GetGeospatialStats(ctx context.Context, req *NearbySearchRequest) (*GeospatialStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			p.category,
			p.province,
			ST_Distance(
				ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
				ST_GeogFromText(ST_AsText(p.location_coordinates))
			) / 1000 as distance_km
		FROM products p
		WHERE p.is_active = true 
		AND p.location_coordinates IS NOT NULL
		AND ST_DWithin(
			ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
			ST_GeogFromText(ST_AsText(p.location_coordinates)),
			$3
		)
		GROUP BY p.category, p.province, distance_km
		ORDER BY distance_km ASC`

	args := []interface{}{req.Longitude, req.Latitude, req.RadiusKm * 1000}

	rows, err := g.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get geospatial stats: %w", err)
	}
	defer rows.Close()

	stats := &GeospatialStats{
		ProductsByCategory: make(map[string]int),
		ProductsByProvince: make(map[string]int),
	}

	var totalDistance float64
	var minDistance, maxDistance float64 = math.MaxFloat64, 0
	var count int

	for rows.Next() {
		var productCount int
		var category, province string
		var distance float64

		err := rows.Scan(&productCount, &category, &province, &distance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats row: %w", err)
		}

		stats.TotalProductsInRadius += productCount
		stats.ProductsByCategory[category] += productCount
		stats.ProductsByProvince[province] += productCount

		totalDistance += distance
		count++

		if distance < minDistance {
			minDistance = distance
		}
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	if count > 0 {
		stats.AverageDistance = totalDistance / float64(count)
	}

	return stats, nil
}

// GetProductsInBounds returns products within a bounding box
func (g *GeospatialService) GetProductsInBounds(ctx context.Context, bounds LocationBounds, category string, limit int) ([]*Product, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.description, p.category, p.subcategory,
			p.price, p.price_type, p.currency, p.unit, p.quantity,
			p.province, p.city, p.seller_name, p.seller_rating,
			p.created_at, p.published_at,
			ST_X(p.location_coordinates) as lng,
			ST_Y(p.location_coordinates) as lat
		FROM products p
		WHERE p.is_active = true 
		AND p.location_coordinates IS NOT NULL
		AND ST_Within(
			p.location_coordinates,
			ST_MakeEnvelope($1, $2, $3, $4, 4326)
		)`

	args := []interface{}{
		bounds.SouthWest.Lng, bounds.SouthWest.Lat, // min lng, min lat
		bounds.NorthEast.Lng, bounds.NorthEast.Lat, // max lng, max lat
	}
	argIndex := 5

	if category != "" {
		query += fmt.Sprintf(" AND p.category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	query += " ORDER BY p.created_at DESC"

	if limit > 0 && limit <= 1000 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		query += " LIMIT 100" // Default limit
	}

	rows, err := g.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get products in bounds: %w", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		var lng, lat sql.NullFloat64

		err := rows.Scan(
			&product.ID, &product.UserID, &product.Title, &product.Description,
			&product.Category, &product.Subcategory, &product.Price, &product.PriceType,
			&product.Currency, &product.Unit, &product.Quantity, &product.Province,
			&product.City, &product.SellerName, &product.SellerRating,
			&product.CreatedAt, &product.PublishedAt, &lng, &lat)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}

		if lng.Valid && lat.Valid {
			product.LocationCoordinates = &Point{
				Lng: lng.Float64,
				Lat: lat.Float64,
			}
		}

		products = append(products, product)
	}

	return products, nil
}

// FindProductsAlongRoute finds products along a route between two points
func (g *GeospatialService) FindProductsAlongRoute(ctx context.Context, start, end Point, corridorWidthKm float64, category string, maxResults int) ([]*NearbyProduct, error) {
	// Create a route line and buffer it to create a corridor
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.description, p.category, p.subcategory,
			p.price, p.price_type, p.currency, p.province, p.city,
			p.seller_name, p.seller_rating, p.created_at,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[0] ELSE NULL END as lng,
			CASE WHEN p.location_coordinates IS NOT NULL THEN p.location_coordinates[1] ELSE NULL END as lat,
			ST_Distance(
				ST_MakeLine(
					ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
					ST_GeogFromText('POINT(' || $3 || ' ' || $4 || ')')
				),
				ST_GeogFromText(ST_AsText(p.location_coordinates))
			) / 1000 as distance_to_route_km
		FROM products p
		WHERE p.is_active = true 
		AND p.location_coordinates IS NOT NULL
		AND ST_DWithin(
			ST_MakeLine(
				ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
				ST_GeogFromText('POINT(' || $3 || ' ' || $4 || ')')
			),
			ST_GeogFromText(ST_AsText(p.location_coordinates)),
			$5
		)`

	args := []interface{}{
		start.Lng, start.Lat,
		end.Lng, end.Lat,
		corridorWidthKm * 1000, // Convert to meters
	}
	argIndex := 6

	if category != "" {
		query += fmt.Sprintf(" AND p.category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	query += " ORDER BY distance_to_route_km ASC"

	if maxResults > 0 && maxResults <= 100 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, maxResults)
	} else {
		query += " LIMIT 50"
	}

	rows, err := g.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find products along route: %w", err)
	}
	defer rows.Close()

	var nearbyProducts []*NearbyProduct
	for rows.Next() {
		product := &Product{}
		var lng, lat, distanceToRoute sql.NullFloat64

		err := rows.Scan(
			&product.ID, &product.UserID, &product.Title, &product.Description,
			&product.Category, &product.Subcategory, &product.Price, &product.PriceType,
			&product.Currency, &product.Province, &product.City, &product.SellerName,
			&product.SellerRating, &product.CreatedAt, &lng, &lat, &distanceToRoute)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}

		if lng.Valid && lat.Valid {
			product.LocationCoordinates = &Point{
				Lng: lng.Float64,
				Lat: lat.Float64,
			}
		}

		nearbyProduct := &NearbyProduct{
			Product:    product,
			DistanceKm: distanceToRoute.Float64,
		}

		nearbyProducts = append(nearbyProducts, nearbyProduct)
	}

	return nearbyProducts, nil
}

// CalculateDistance calculates the distance between two points in kilometers
func (g *GeospatialService) CalculateDistance(ctx context.Context, point1, point2 Point) (float64, error) {
	query := `
		SELECT ST_Distance(
			ST_GeogFromText('POINT(' || $1 || ' ' || $2 || ')'),
			ST_GeogFromText('POINT(' || $3 || ' ' || $4 || ')')
		) / 1000 as distance_km`

	var distance float64
	err := g.db.QueryRowContext(ctx, query, point1.Lng, point1.Lat, point2.Lng, point2.Lat).Scan(&distance)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate distance: %w", err)
	}

	return distance, nil
}

// GetProductDensityGrid returns a grid showing product density in an area
func (g *GeospatialService) GetProductDensityGrid(ctx context.Context, bounds LocationBounds, gridSize int, category string) ([][]int, error) {
	// This is a simplified implementation - in production you might use more sophisticated gridding
	lngStep := (bounds.NorthEast.Lng - bounds.SouthWest.Lng) / float64(gridSize)
	latStep := (bounds.NorthEast.Lat - bounds.SouthWest.Lat) / float64(gridSize)

	grid := make([][]int, gridSize)
	for i := range grid {
		grid[i] = make([]int, gridSize)
	}

	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			minLng := bounds.SouthWest.Lng + float64(j)*lngStep
			maxLng := bounds.SouthWest.Lng + float64(j+1)*lngStep
			minLat := bounds.SouthWest.Lat + float64(i)*latStep
			maxLat := bounds.SouthWest.Lat + float64(i+1)*latStep

			query := `
				SELECT COUNT(*)
				FROM products p
				WHERE p.is_active = true 
						AND p.location_coordinates IS NOT NULL
				AND ST_Within(
					p.location_coordinates,
					ST_MakeEnvelope($1, $2, $3, $4, 4326)
				)`

			args := []interface{}{minLng, minLat, maxLng, maxLat}

			if category != "" {
				query += " AND p.category = $5"
				args = append(args, category)
			}

			var count int
			err := g.db.QueryRowContext(ctx, query, args...).Scan(&count)
			if err != nil {
				return nil, fmt.Errorf("failed to get grid cell count: %w", err)
			}

			grid[i][j] = count
		}
	}

	return grid, nil
}

// Helper function to create geospatial search handlers
func (g *GeospatialService) RegisterRoutes(router *gin.RouterGroup) {
	geo := router.Group("/geo")
	{
		geo.POST("/nearby", g.handleNearbySearch)
		geo.POST("/bounds", g.handleBoundsSearch)
		geo.POST("/route", g.handleRouteSearch)
		geo.POST("/stats", g.handleGeospatialStats)
		geo.POST("/distance", g.handleDistanceCalculation)
		geo.POST("/density", g.handleDensityGrid)
	}
}

// Handler implementations
func (g *GeospatialService) handleNearbySearch(c *gin.Context) {
	var req NearbySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	results, err := g.FindNearbyProducts(c.Request.Context(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(200, results)
}

func (g *GeospatialService) handleBoundsSearch(c *gin.Context) {
	// Simple bounds search using nearby search as base
	var req NearbySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	results, err := g.FindNearbyProducts(c.Request.Context(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(200, results)
}

func (g *GeospatialService) handleRouteSearch(c *gin.Context) {
	type RouteRequest struct {
		Start      Point   `json:"start"`
		End        Point   `json:"end"`
		BufferKm   float64 `json:"buffer_km"`
		Category   string  `json:"category"`
		MaxResults int     `json:"max_results"`
	}

	var req RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	results, err := g.FindProductsAlongRoute(c.Request.Context(), req.Start, req.End, req.BufferKm, req.Category, req.MaxResults)
	if err != nil {
		c.JSON(500, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(200, results)
}

func (g *GeospatialService) handleGeospatialStats(c *gin.Context) {
	var req NearbySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	stats, err := g.GetGeospatialStats(c.Request.Context(), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": "Stats failed"})
		return
	}

	c.JSON(200, stats)
}

func (g *GeospatialService) handleDistanceCalculation(c *gin.Context) {
	type DistanceRequest struct {
		Point1 Point `json:"point1"`
		Point2 Point `json:"point2"`
	}

	var req DistanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	distance, err := g.CalculateDistance(c.Request.Context(), req.Point1, req.Point2)
	if err != nil {
		c.JSON(500, gin.H{"error": "Calculation failed"})
		return
	}

	c.JSON(200, gin.H{"distance_km": distance})
}

func (g *GeospatialService) handleDensityGrid(c *gin.Context) {
	// Simplified density response
	var req NearbySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Return a basic response for density grid
	c.JSON(200, gin.H{"message": "Density grid not implemented"})
}
