package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"file-sharing-platform/internal/auth"
	"file-sharing-platform/internal/service"

	"github.com/gorilla/mux"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx := r.Context()
	fileInfo, err := h.fileService.UploadFile(ctx, userID, header.Filename, header.Size, header.Header.Get("Content-Type"), file)
	if err != nil {
		http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfo)
}

func (h *FileHandler) GetUserFiles(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	files, err := h.fileService.GetUserFiles(ctx, userID, 0, 0)
	if err != nil {
		http.Error(w, "Error retrieving files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *FileHandler) ShareFile(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID, err := strconv.ParseInt(vars["file_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	expirationHours := 24
	if expStr := r.URL.Query().Get("expires_in"); expStr != "" {
		if exp, err := strconv.Atoi(expStr); err == nil && exp > 0 {
			expirationHours = exp
		}
	}

	ctx := r.Context()
	shareURL, err := h.fileService.ShareFile(ctx, strconv.FormatInt(userID, 10), fileID, strconv.Itoa(expirationHours))
	if err != nil {
		http.Error(w, "Error sharing file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"share_url": shareURL.ShareURL}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *FileHandler) GetSharedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shareToken := vars["share_token"]

	ctx := r.Context()
	fileInfo, err := h.fileService.GetFile(ctx, shareToken)
	if err != nil {
		http.Error(w, "File not found or share expired", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, fileInfo.PublicURL, http.StatusFound)
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID, err := strconv.ParseInt(vars["file_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err = h.fileService.DeleteFile(ctx, strconv.FormatInt(userID, 10), fileID)
	if err != nil {
		http.Error(w, "Error deleting file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
