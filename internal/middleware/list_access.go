package middleware

import (
	"net/http"
	"strconv"

	"family-shopping-list-api/internal/member/repository"
	"family-shopping-list-api/pkg/response"
)

func RequireListMember(repo *repository.Repository, ownerOnly bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID, err := strconv.ParseUint(r.PathValue("list_id"), 10, 64)
		if err != nil || listID == 0 {
			response.Error(w, http.StatusBadRequest, "bad_request", "无效的清单 ID")
			return
		}
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "缺少用户身份")
			return
		}

		var allowed bool
		if ownerOnly {
			allowed, err = repo.IsOwner(r.Context(), listID, userID)
		} else {
			allowed, err = repo.IsMember(r.Context(), listID, userID)
		}
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "authorization_failed", err.Error())
			return
		}
		if !allowed {
			response.Error(w, http.StatusForbidden, "forbidden", "没有访问该清单的权限")
			return
		}
		next(w, r)
	}
}
