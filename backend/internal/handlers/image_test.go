package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"recipe-backend/internal/config"
	"recipe-backend/internal/database"
	"recipe-backend/internal/models"
	"recipe-backend/internal/repository"
)

type fakeObjectStore struct {
	putContentType string
	putCount       int
	deleteCount    int
}

func (f *fakeObjectStore) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.putCount++
	f.putContentType = *input.ContentType
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeObjectStore) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.deleteCount++
	return &s3.DeleteObjectOutput{}, nil
}

func setupImageHandler(t *testing.T) (*RecipeHandler, *fakeObjectStore, int64) {
	t.Helper()

	db, err := database.New(filepath.Join(t.TempDir(), "images.db"))
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() { database.Close(db) })
	if err := database.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	repo := repository.NewRecipeRepository(db)
	recipe, err := repo.Create(models.CreateRecipeRequest{
		Title:       "Image Recipe",
		Categories:  []string{"Dinner"},
		Ingredients: []string{"salt"},
		Steps:       []string{"cook"},
		Links:       []models.Link{},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	store := &fakeObjectStore{}
	h := NewRecipeHandler(repo, &config.Config{
		R2BucketName: "bucket",
		R2PublicURL:  "https://cdn.example",
	})
	h.s3Client = store

	return h, store, recipe.ID
}

func TestUploadRecipeImageAllowsPNGAndSavesURL(t *testing.T) {
	h, store, id := setupImageHandler(t)
	body, contentType := multipartBody(t, "image", "photo.png", []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d,
	})

	req := httptest.NewRequest(http.MethodPost, "/recipes/1/image", body)
	req.Header.Set("Content-Type", contentType)
	req = withRecipeID(req, id)
	rr := httptest.NewRecorder()

	h.UploadRecipeImage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rr.Code, rr.Body.String())
	}
	if store.putCount != 1 || store.putContentType != "image/png" {
		t.Fatalf("unexpected put state: count=%d type=%q", store.putCount, store.putContentType)
	}
}

func TestUploadRecipeImageRejectsUnsupportedImageType(t *testing.T) {
	h, store, id := setupImageHandler(t)
	body, contentType := multipartBody(t, "image", "photo.gif", []byte("GIF87a"))

	req := httptest.NewRequest(http.MethodPost, "/recipes/1/image", body)
	req.Header.Set("Content-Type", contentType)
	req = withRecipeID(req, id)
	rr := httptest.NewRecorder()

	h.UploadRecipeImage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body %s", rr.Code, rr.Body.String())
	}
	if store.putCount != 0 {
		t.Fatalf("object store should not be called")
	}
}

func multipartBody(t *testing.T, field, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close multipart: %v", err)
	}
	return body, writer.FormDataContentType()
}

func withRecipeID(req *http.Request, id int64) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(id, 10))
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}
