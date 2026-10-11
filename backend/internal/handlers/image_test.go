package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	putKeys        []string
	deleteKeys     []string
	cacheControls  []string
	// onPut runs after each stored object, before PutObject returns.
	onPut func(key string)
}

func (f *fakeObjectStore) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.putCount++
	f.putContentType = *input.ContentType
	f.putKeys = append(f.putKeys, *input.Key)
	f.cacheControls = append(f.cacheControls, aws.ToString(input.CacheControl))
	if f.onPut != nil {
		f.onPut(*input.Key)
	}
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeObjectStore) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.deleteCount++
	f.deleteKeys = append(f.deleteKeys, *input.Key)
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

type filePart struct {
	field   string
	content []byte
}

var jpegBytes = []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}

func multipartParts(t *testing.T, parts ...filePart) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, p := range parts {
		part, err := writer.CreateFormFile(p.field, p.field+".jpg")
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := part.Write(p.content); err != nil {
			t.Fatalf("write part: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close multipart: %v", err)
	}
	return body, writer.FormDataContentType()
}

func postImage(t *testing.T, h *RecipeHandler, id int64, parts ...filePart) *httptest.ResponseRecorder {
	body, contentType := multipartParts(t, parts...)
	req := httptest.NewRequest(http.MethodPost, "/recipes/1/image", body)
	req.Header.Set("Content-Type", contentType)
	req = withRecipeID(req, id)
	rr := httptest.NewRecorder()
	h.UploadRecipeImage(rr, req)
	return rr
}

func allThreeParts() []filePart {
	return []filePart{{"image", jpegBytes}, {"thumb", jpegBytes}, {"original", jpegBytes}}
}

func TestUploadRecipeImageStoresFramedThumbAndOriginal(t *testing.T) {
	h, store, id := setupImageHandler(t)

	rr := postImage(t, h, id, allThreeParts()...)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rr.Code, rr.Body.String())
	}
	wantKeys := []string{"recipes/1", "recipes/1-thumb", "recipes/1-original"}
	if !reflect.DeepEqual(store.putKeys, wantKeys) {
		t.Fatalf("put keys = %v, want %v", store.putKeys, wantKeys)
	}
	for _, cc := range store.cacheControls {
		if !strings.Contains(cc, "immutable") {
			t.Fatalf("missing Cache-Control on an object: %v", store.cacheControls)
		}
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for field, key := range map[string]string{
		"image_url":          "recipes/1?v=",
		"image_thumb_url":    "recipes/1-thumb?v=",
		"image_original_url": "recipes/1-original?v=",
	} {
		if !strings.HasPrefix(resp[field], "https://cdn.example/"+key) {
			t.Fatalf("%s = %q, want prefix %q", field, resp[field], "https://cdn.example/"+key)
		}
	}

	recipe, err := h.repo.FindByID(id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if recipe.ImageURL == nil || recipe.ImageThumbURL == nil || recipe.ImageOriginalURL == nil {
		t.Fatalf("expected all image URLs saved, got %+v", recipe)
	}
}

func TestUploadRecipeImageRequiresThumbAndOriginalTogether(t *testing.T) {
	h, store, id := setupImageHandler(t)

	rr := postImage(t, h, id, filePart{"image", jpegBytes}, filePart{"thumb", jpegBytes})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body %s", rr.Code, rr.Body.String())
	}
	if store.putCount != 0 {
		t.Fatalf("nothing should be stored, got %v", store.putKeys)
	}
}

func TestUploadRecipeImageRejectsBadThumbType(t *testing.T) {
	h, store, id := setupImageHandler(t)

	rr := postImage(t, h, id,
		filePart{"image", jpegBytes}, filePart{"thumb", []byte("GIF87a")}, filePart{"original", jpegBytes})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body %s", rr.Code, rr.Body.String())
	}
	if store.putCount != 0 {
		t.Fatalf("nothing should be stored, got %v", store.putKeys)
	}
}

func TestUploadRecipeImageRejectsPartOver15MB(t *testing.T) {
	h, store, id := setupImageHandler(t)
	big := append(append([]byte{}, jpegBytes...), make([]byte, 15<<20)...)

	rr := postImage(t, h, id, filePart{"image", big})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body %s", rr.Code, rr.Body.String())
	}
	if store.putCount != 0 {
		t.Fatalf("nothing should be stored, got %v", store.putKeys)
	}
}

func TestUploadRecipeImageRemovesObjectsWhenRecipeVanishes(t *testing.T) {
	h, store, id := setupImageHandler(t)
	store.onPut = func(key string) {
		if key == "recipes/1-original" {
			if err := h.repo.Delete(id); err != nil {
				t.Errorf("Delete recipe: %v", err)
			}
		}
	}

	rr := postImage(t, h, id, allThreeParts()...)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body %s", rr.Code, rr.Body.String())
	}
	want := []string{"recipes/1", "recipes/1-thumb", "recipes/1-original"}
	if !reflect.DeepEqual(store.deleteKeys, want) {
		t.Fatalf("deleted keys = %v, want %v", store.deleteKeys, want)
	}
}

func TestPlainUploadReplacesFramedImageAndRemovesStaleVariants(t *testing.T) {
	h, store, id := setupImageHandler(t)
	if rr := postImage(t, h, id, allThreeParts()...); rr.Code != http.StatusOK {
		t.Fatalf("setup upload: %d %s", rr.Code, rr.Body.String())
	}

	rr := postImage(t, h, id, filePart{"image", jpegBytes})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rr.Code, rr.Body.String())
	}
	want := []string{"recipes/1-thumb", "recipes/1-original"}
	if !reflect.DeepEqual(store.deleteKeys, want) {
		t.Fatalf("deleted keys = %v, want %v", store.deleteKeys, want)
	}
	recipe, _ := h.repo.FindByID(id)
	if recipe.ImageURL == nil || recipe.ImageThumbURL != nil || recipe.ImageOriginalURL != nil {
		t.Fatalf("expected only image_url set, got %+v", recipe)
	}
}

func TestDeleteRecipeImageRemovesAllVariants(t *testing.T) {
	h, store, id := setupImageHandler(t)
	if rr := postImage(t, h, id, allThreeParts()...); rr.Code != http.StatusOK {
		t.Fatalf("setup upload: %d %s", rr.Code, rr.Body.String())
	}

	req := withRecipeID(httptest.NewRequest(http.MethodDelete, "/recipes/1/image", nil), id)
	rr := httptest.NewRecorder()
	h.DeleteRecipeImage(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body %s", rr.Code, rr.Body.String())
	}
	want := []string{"recipes/1", "recipes/1-thumb", "recipes/1-original"}
	if !reflect.DeepEqual(store.deleteKeys, want) {
		t.Fatalf("deleted keys = %v, want %v", store.deleteKeys, want)
	}
	recipe, _ := h.repo.FindByID(id)
	if recipe.ImageURL != nil || recipe.ImageThumbURL != nil || recipe.ImageOriginalURL != nil {
		t.Fatalf("expected all image URLs cleared, got %+v", recipe)
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
