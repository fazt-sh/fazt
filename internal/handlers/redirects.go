package handlers

import (
	"net/http"
	"strconv"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/database"
)

// DeleteRedirectHandler handles DELETE /api/redirects/{id}
func DeleteRedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		api.BadRequest(w, "Redirect ID required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.BadRequest(w, "Invalid redirect ID")
		return
	}

	// Delete from database
	db := database.GetDB()
	result, err := db.Exec("DELETE FROM redirects WHERE id = ?", id)
	if err != nil {
		api.ServerError(w, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		api.NotFound(w, "Redirect not found")
		return
	}

	api.JSON(w, http.StatusOK, map[string]string{"message": "Redirect deleted"}, nil)
}
