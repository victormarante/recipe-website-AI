package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"recipe-backend/internal/repository"
	"recipe-backend/internal/respond"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadRecipeImage handles POST /api/v1/recipes/{id}/image
func (h *RecipeHandler) UploadRecipeImage(w http.ResponseWriter, r *http.Request) {
	if h.s3Client == nil {
		respond.Error(w, http.StatusServiceUnavailable, "Image storage not configured")
		return
	}

	id, err := parseID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	if _, err := h.repo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to fetch recipe", err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		respond.Error(w, http.StatusBadRequest, "File too large or invalid multipart form")
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Missing 'image' field in form")
		return
	}
	defer file.Close()

	imgBytes, err := io.ReadAll(file)
	if err != nil {
		respond.InternalError(w, "Failed to read file", err)
		return
	}

	contentType := http.DetectContentType(imgBytes)
	if !allowedImageTypes[contentType] {
		respond.Error(w, http.StatusBadRequest, "File must be a JPEG, PNG, or WebP image")
		return
	}

	key := fmt.Sprintf("recipes/%d", id)
	log.Printf("info: uploading image to R2 bucket=%s key=%s size=%d type=%s", h.cfg.R2BucketName, key, len(imgBytes), contentType)
	if _, err = h.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:        aws.String(h.cfg.R2BucketName),
		Key:           aws.String(key),
		Body:          bytes.NewReader(imgBytes),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(imgBytes))),
	}); err != nil {
		respond.InternalError(w, "Failed to upload image", err)
		return
	}
	log.Printf("info: image uploaded successfully key=%s", key)

	imageURL := strings.TrimRight(h.cfg.R2PublicURL, "/") + "/" + key
	if err := h.repo.UpdateImageURL(id, &imageURL); err != nil {
		if _, deleteErr := h.s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(h.cfg.R2BucketName),
			Key:    aws.String(key),
		}); deleteErr != nil {
			log.Printf("error: cleanup uploaded image after database failure key=%s: %v", key, deleteErr)
		}
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to save image URL", err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"image_url": imageURL})
}

// DeleteRecipeImage handles DELETE /api/v1/recipes/{id}/image
func (h *RecipeHandler) DeleteRecipeImage(w http.ResponseWriter, r *http.Request) {
	if h.s3Client == nil {
		respond.Error(w, http.StatusServiceUnavailable, "Image storage not configured")
		return
	}

	id, err := parseID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	recipe, err := h.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to fetch recipe", err)
		return
	}

	if recipe.ImageURL != nil {
		key := fmt.Sprintf("recipes/%d", id)
		if _, err = h.s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(h.cfg.R2BucketName),
			Key:    aws.String(key),
		}); err != nil {
			respond.InternalError(w, "Failed to delete image from storage", err)
			return
		}
	}

	if err := h.repo.UpdateImageURL(id, nil); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to clear image URL", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
