package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

const (
	// maxImagePartBytes caps each uploaded file (image, thumb, original).
	maxImagePartBytes = 15 << 20
	// maxImageFormMemory is how much of the multipart form is held in memory;
	// larger parts spill to temporary files.
	maxImageFormMemory = 10 << 20
	// imageCacheControl lets browsers keep images for a year. Re-uploads get a
	// new ?v= query on their URL, so a replaced image is never served stale.
	imageCacheControl = "public, max-age=31536000, immutable"
)

var (
	errImagePartTooLarge = errors.New("image part too large")
	errImageType         = errors.New("must be a JPEG, PNG, or WebP image")
)

// imageKeys returns the object keys for a recipe's framed image, card
// thumbnail and untouched original.
func imageKeys(id int64) (image, thumb, original string) {
	image = fmt.Sprintf("recipes/%d", id)
	return image, image + "-thumb", image + "-original"
}

// imageUpload is one validated file ready to be stored.
type imageUpload struct {
	key         string
	data        []byte
	contentType string
}

// readImagePart reads an optional multipart image field. It returns nil data
// when the field is absent.
func readImagePart(r *http.Request, field string) ([]byte, string, error) {
	file, _, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxImagePartBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxImagePartBytes {
		return nil, "", errImagePartTooLarge
	}
	contentType := http.DetectContentType(data)
	if !allowedImageTypes[contentType] {
		return nil, "", fmt.Errorf("%s %w", field, errImageType)
	}
	return data, contentType, nil
}

// deleteObjects removes keys from the bucket. It returns the first error but
// still tries every key; deleting a missing key is not an error on S3/R2.
func (h *RecipeHandler) deleteObjects(ctx context.Context, keys ...string) error {
	var firstErr error
	for _, key := range keys {
		if _, err := h.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(h.cfg.R2BucketName),
			Key:    aws.String(key),
		}); err != nil {
			log.Printf("error: delete image object key=%s: %v", key, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// UploadRecipeImage handles POST /api/v1/recipes/{id}/image
//
// The "image" field is required. The frontend also sends "thumb" (a small copy
// for list cards) and "original" (the untouched upload, kept as a backup); the
// two must be sent together. With only "image" the file is stored as-is.
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

	existing, err := h.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to fetch recipe", err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 3*maxImagePartBytes+(1<<20))
	if err := r.ParseMultipartForm(maxImageFormMemory); err != nil {
		respond.Error(w, http.StatusBadRequest, "File too large or invalid multipart form")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	imageKey, thumbKey, originalKey := imageKeys(id)
	var uploads []imageUpload
	for _, part := range []struct{ field, key string }{
		{"image", imageKey}, {"thumb", thumbKey}, {"original", originalKey},
	} {
		data, contentType, err := readImagePart(r, part.field)
		switch {
		case errors.Is(err, errImagePartTooLarge):
			respond.Error(w, http.StatusBadRequest, "File too large (max 15 MB)")
			return
		case errors.Is(err, errImageType):
			respond.Error(w, http.StatusBadRequest, err.Error())
			return
		case err != nil:
			respond.InternalError(w, "Failed to read file", err)
			return
		case data == nil && part.field == "image":
			respond.Error(w, http.StatusBadRequest, "Missing 'image' field in form")
			return
		case data == nil:
			continue
		}
		uploads = append(uploads, imageUpload{key: part.key, data: data, contentType: contentType})
	}
	if len(uploads) == 2 {
		respond.Error(w, http.StatusBadRequest, "'thumb' and 'original' must be sent together")
		return
	}

	var uploaded []string
	for _, u := range uploads {
		log.Printf("info: uploading image to R2 bucket=%s key=%s size=%d type=%s", h.cfg.R2BucketName, u.key, len(u.data), u.contentType)
		if _, err := h.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:        aws.String(h.cfg.R2BucketName),
			Key:           aws.String(u.key),
			Body:          bytes.NewReader(u.data),
			ContentType:   aws.String(u.contentType),
			ContentLength: aws.Int64(int64(len(u.data))),
			CacheControl:  aws.String(imageCacheControl),
		}); err != nil {
			h.deleteObjects(r.Context(), uploaded...)
			respond.InternalError(w, "Failed to upload image", err)
			return
		}
		uploaded = append(uploaded, u.key)
	}
	log.Printf("info: image uploaded successfully recipe=%d files=%d", id, len(uploaded))

	baseURL := strings.TrimRight(h.cfg.R2PublicURL, "/") + "/"
	version := "?v=" + strconv.FormatInt(time.Now().Unix(), 10)
	imageURL := baseURL + imageKey + version
	var thumbURL, originalURL *string
	response := map[string]string{"image_url": imageURL}
	if len(uploads) == 3 {
		t, o := baseURL+thumbKey+version, baseURL+originalKey+version
		thumbURL, originalURL = &t, &o
		response["image_thumb_url"] = t
		response["image_original_url"] = o
	}

	if err := h.repo.UpdateImages(id, &imageURL, thumbURL, originalURL); err != nil {
		h.deleteObjects(r.Context(), uploaded...)
		if errors.Is(err, repository.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respond.InternalError(w, "Failed to save image URL", err)
		return
	}

	// A plain upload replaces a recipe that had a thumbnail and original, so
	// their objects are now stale.
	if len(uploads) == 1 && (existing.ImageThumbURL != nil || existing.ImageOriginalURL != nil) {
		h.deleteObjects(r.Context(), thumbKey, originalKey)
	}

	respond.JSON(w, http.StatusOK, response)
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
		imageKey, thumbKey, originalKey := imageKeys(id)
		if err := h.deleteObjects(r.Context(), imageKey, thumbKey, originalKey); err != nil {
			respond.InternalError(w, "Failed to delete image from storage", err)
			return
		}
	}

	if err := h.repo.UpdateImages(id, nil, nil, nil); err != nil {
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
