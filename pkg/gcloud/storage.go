package gcloud

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type StorageClient struct {
	client *storage.Client
	bucket string
}

type UploadResult struct {
	URL           string    `json:"url"`
	StoragePath   string    `json:"storage_path"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size"`
	MimeType      string    `json:"mime_type"`
	UploadedAt    time.Time `json:"uploaded_at"`
}

type UploadOptions struct {
	Directory   string            `json:"directory"`   // e.g., "products", "users", "documents"
	SubDirectory string           `json:"sub_directory"` // e.g., user ID, product ID
	FileName    string            `json:"file_name"`   // custom filename, if empty will use original
	Metadata    map[string]string `json:"metadata"`
	PublicRead  bool              `json:"public_read"` // whether file should be publicly readable
	CacheControl string           `json:"cache_control"` // cache control header
}

func NewStorageClient(ctx context.Context, projectID, credentialsFile, bucketName string) (*StorageClient, error) {
	var client *storage.Client
	var err error

	if credentialsFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	} else {
		// Use default credentials (for Cloud Run deployment)
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &StorageClient{
		client: client,
		bucket: bucketName,
	}, nil
}

func NewStorageClientWithEmulator(ctx context.Context, projectID, bucketName, emulatorHost string) (*StorageClient, error) {
	// Set the emulator host environment variable
	os.Setenv("STORAGE_EMULATOR_HOST", emulatorHost)
	
	// Create client without credentials for emulator
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create emulator storage client: %w", err)
	}

	// Create bucket if it doesn't exist (emulator starts empty)
	bucket := client.Bucket(bucketName)
	if err := bucket.Create(ctx, projectID, nil); err != nil {
		// Ignore error if bucket already exists
		fmt.Printf("Bucket creation result (may already exist): %v\n", err)
	}

	return &StorageClient{
		client: client,
		bucket: bucketName,
	}, nil
}

func (sc *StorageClient) Close() error {
	return sc.client.Close()
}

// UploadFile uploads a file to Google Cloud Storage
func (sc *StorageClient) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, options UploadOptions) (*UploadResult, error) {
	// Generate storage path
	storagePath := sc.generateStoragePath(header.Filename, options)
	
	// Create object
	obj := sc.client.Bucket(sc.bucket).Object(storagePath)
	writer := obj.NewWriter(ctx)

	// Set metadata
	writer.ObjectAttrs.ContentType = header.Header.Get("Content-Type")
	if writer.ObjectAttrs.ContentType == "" {
		writer.ObjectAttrs.ContentType = sc.detectContentType(header.Filename)
	}

	// Set cache control
	if options.CacheControl != "" {
		writer.ObjectAttrs.CacheControl = options.CacheControl
	} else {
		writer.ObjectAttrs.CacheControl = "public, max-age=86400" // 1 day default
	}

	// Set custom metadata
	if options.Metadata != nil {
		writer.ObjectAttrs.Metadata = options.Metadata
	}

	// Copy file content
	fileSize, err := io.Copy(writer, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Make object public if requested (skip for emulator)
	if options.PublicRead && os.Getenv("STORAGE_EMULATOR_HOST") == "" {
		if err := sc.makeObjectPublic(ctx, storagePath); err != nil {
			return nil, fmt.Errorf("failed to make object public: %w", err)
		}
	}

	// Generate public URL
	url := sc.generatePublicURL(storagePath)

	return &UploadResult{
		URL:         url,
		StoragePath: storagePath,
		FileName:    header.Filename,
		FileSize:    fileSize,
		MimeType:    writer.ObjectAttrs.ContentType,
		UploadedAt:  time.Now(),
	}, nil
}

// UploadFileFromBytes uploads a file from byte slice
func (sc *StorageClient) UploadFileFromBytes(ctx context.Context, data []byte, fileName, contentType string, options UploadOptions) (*UploadResult, error) {
	// Generate storage path
	storagePath := sc.generateStoragePath(fileName, options)
	
	// Create object
	obj := sc.client.Bucket(sc.bucket).Object(storagePath)
	writer := obj.NewWriter(ctx)

	// Set metadata
	writer.ObjectAttrs.ContentType = contentType
	if writer.ObjectAttrs.ContentType == "" {
		writer.ObjectAttrs.ContentType = sc.detectContentType(fileName)
	}

	// Set cache control
	if options.CacheControl != "" {
		writer.ObjectAttrs.CacheControl = options.CacheControl
	} else {
		writer.ObjectAttrs.CacheControl = "public, max-age=86400" // 1 day default
	}

	// Set custom metadata
	if options.Metadata != nil {
		writer.ObjectAttrs.Metadata = options.Metadata
	}

	// Write data
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Make object public if requested (skip for emulator)
	if options.PublicRead && os.Getenv("STORAGE_EMULATOR_HOST") == "" {
		if err := sc.makeObjectPublic(ctx, storagePath); err != nil {
			return nil, fmt.Errorf("failed to make object public: %w", err)
		}
	}

	// Generate public URL
	url := sc.generatePublicURL(storagePath)

	return &UploadResult{
		URL:         url,
		StoragePath: storagePath,
		FileName:    fileName,
		FileSize:    int64(len(data)),
		MimeType:    writer.ObjectAttrs.ContentType,
		UploadedAt:  time.Now(),
	}, nil
}

// DeleteFile deletes a file from Google Cloud Storage
func (sc *StorageClient) DeleteFile(ctx context.Context, storagePath string) error {
	obj := sc.client.Bucket(sc.bucket).Object(storagePath)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object %s: %w", storagePath, err)
	}
	return nil
}

// GetFileURL generates a signed URL for private files or public URL for public files
func (sc *StorageClient) GetFileURL(ctx context.Context, storagePath string, expiration time.Duration) (string, error) {
	obj := sc.client.Bucket(sc.bucket).Object(storagePath)
	
	// Check if object exists and is public
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get object attributes: %w", err)
	}

	// If object is public, return public URL
	for _, acl := range attrs.ACL {
		if acl.Entity == storage.AllUsers && acl.Role == storage.RoleReader {
			return sc.generatePublicURL(storagePath), nil
		}
	}

	// Generate signed URL for private objects
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	signedURL, err := storage.SignedURL(sc.bucket, storagePath, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return signedURL, nil
}

// ListFiles lists files in a directory
func (sc *StorageClient) ListFiles(ctx context.Context, prefix string, maxResults int) ([]*storage.ObjectAttrs, error) {
	query := &storage.Query{
		Prefix: prefix,
	}

	it := sc.client.Bucket(sc.bucket).Objects(ctx, query)
	
	var objects []*storage.ObjectAttrs
	count := 0
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate objects: %w", err)
		}
		
		objects = append(objects, attrs)
		count++
		
		// Check if we've reached the maximum results limit
		if maxResults > 0 && count >= maxResults {
			break
		}
	}

	return objects, nil
}

// GenerateResizedImageURL generates a URL for a resized image using Cloud Storage's image serving
func (sc *StorageClient) GenerateResizedImageURL(storagePath string, width, height int, quality int) string {
	baseURL := sc.generatePublicURL(storagePath)
	
	// Add image transformation parameters
	params := []string{}
	if width > 0 {
		params = append(params, fmt.Sprintf("w%d", width))
	}
	if height > 0 {
		params = append(params, fmt.Sprintf("h%d", height))
	}
	if quality > 0 && quality <= 100 {
		params = append(params, fmt.Sprintf("q%d", quality))
	}

	if len(params) > 0 {
		return fmt.Sprintf("%s=s%s", baseURL, strings.Join(params, "-"))
	}

	return baseURL
}

// Helper functions
func (sc *StorageClient) generateStoragePath(fileName string, options UploadOptions) string {
	timestamp := time.Now().Format("20060102_150405")
	
	// Clean filename
	cleanFileName := strings.ReplaceAll(fileName, " ", "_")
	cleanFileName = strings.ToLower(cleanFileName)
	
	// Use custom filename if provided
	if options.FileName != "" {
		ext := filepath.Ext(cleanFileName)
		cleanFileName = options.FileName + ext
	}

	// Add timestamp to prevent collisions
	name := strings.TrimSuffix(cleanFileName, filepath.Ext(cleanFileName))
	ext := filepath.Ext(cleanFileName)
	timestampedFileName := fmt.Sprintf("%s_%s%s", name, timestamp, ext)

	// Build path
	pathParts := []string{}
	if options.Directory != "" {
		pathParts = append(pathParts, options.Directory)
	}
	if options.SubDirectory != "" {
		pathParts = append(pathParts, options.SubDirectory)
	}
	pathParts = append(pathParts, timestampedFileName)

	return strings.Join(pathParts, "/")
}

func (sc *StorageClient) generatePublicURL(storagePath string) string {
	// Use emulator URL if STORAGE_EMULATOR_HOST is set
	if emulatorHost := os.Getenv("STORAGE_EMULATOR_HOST"); emulatorHost != "" {
		// URL-encode the storage path for emulator API
		encodedPath := strings.ReplaceAll(storagePath, "/", "%2F")
		// Remove http:// prefix if present since we add it in the format string
		cleanHost := strings.TrimPrefix(emulatorHost, "http://")
		cleanHost = strings.TrimPrefix(cleanHost, "https://")
		return fmt.Sprintf("http://%s/download/storage/v1/b/%s/o/%s?alt=media", cleanHost, sc.bucket, encodedPath)
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", sc.bucket, storagePath)
}

func (sc *StorageClient) makeObjectPublic(ctx context.Context, storagePath string) error {
	obj := sc.client.Bucket(sc.bucket).Object(storagePath)
	acl := obj.ACL()
	
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return fmt.Errorf("failed to set ACL: %w", err)
	}
	
	return nil
}

func (sc *StorageClient) detectContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".json": "application/json",
	}

	if contentType, exists := mimeTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}

// ValidateImageFile validates if the uploaded file is a valid image
func ValidateImageFile(header *multipart.FileHeader) error {
	// Check file size (max 10MB for images)
	const maxSize = 10 * 1024 * 1024 // 10MB
	if header.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	allowedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("invalid image type: %s. Allowed types: %v", contentType, allowedTypes)
}

// ValidateDocumentFile validates if the uploaded file is a valid document
func ValidateDocumentFile(header *multipart.FileHeader) error {
	// Check file size (max 50MB for documents)
	const maxSize = 50 * 1024 * 1024 // 50MB
	if header.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	allowedTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
		"text/csv",
		"application/json",
	}

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("invalid document type: %s. Allowed types: %v", contentType, allowedTypes)
}