package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
)

// UploadRecipeImage handles POST /api/v1/recipes/{id}/image
func (h *RecipeHandler) UploadRecipeImage(w http.ResponseWriter, r *http.Request) {
	if h.s3Client == nil {
		h.sendError(w, http.StatusServiceUnavailable, "Image storage not configured", fmt.Errorf("R2 credentials not set"))
		return
	}

	id, err := parseID(r)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid recipe ID", err)
		return
	}

	if _, err := h.repo.FindByID(id); err != nil {
		h.sendError(w, http.StatusNotFound, "Recipe not found", err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		h.sendError(w, http.StatusBadRequest, "File too large or invalid multipart form", err)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Missing 'image' field in form", err)
		return
	}
	defer file.Close()

	imgBytes, err := io.ReadAll(file)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to read file", err)
		return
	}

	contentType := http.DetectContentType(imgBytes)
	if !strings.HasPrefix(contentType, "image/") {
		h.sendError(w, http.StatusBadRequest, "File must be an image", fmt.Errorf("detected type: %s", contentType))
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
		h.sendError(w, http.StatusInternalServerError, "Failed to upload image", err)
		return
	}
	log.Printf("info: image uploaded successfully key=%s", key)

	imageURL := strings.TrimRight(h.cfg.R2PublicURL, "/") + "/" + key
	if err := h.repo.UpdateImageURL(id, &imageURL); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to save image URL", err)
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]string{"image_url": imageURL})
}

// DeleteRecipeImage handles DELETE /api/v1/recipes/{id}/image
func (h *RecipeHandler) DeleteRecipeImage(w http.ResponseWriter, r *http.Request) {
	if h.s3Client == nil {
		h.sendError(w, http.StatusServiceUnavailable, "Image storage not configured", fmt.Errorf("R2 credentials not set"))
		return
	}

	id, err := parseID(r)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid recipe ID", err)
		return
	}

	recipe, err := h.repo.FindByID(id)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Recipe not found", err)
		return
	}

	if recipe.ImageURL != nil {
		key := fmt.Sprintf("recipes/%d", id)
		if _, err = h.s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(h.cfg.R2BucketName),
			Key:    aws.String(key),
		}); err != nil {
			h.sendError(w, http.StatusInternalServerError, "Failed to delete image from storage", err)
			return
		}
	}

	if err := h.repo.UpdateImageURL(id, nil); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to clear image URL", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
