package api

import (
	"database/sql"
	"net/http"

	"family-shopping-list-api/internal/config"
	"family-shopping-list-api/internal/middleware"

	inviteHandler "family-shopping-list-api/internal/invite/handler"
	inviteRepo "family-shopping-list-api/internal/invite/repository"
	inviteService "family-shopping-list-api/internal/invite/service"
	itemHandler "family-shopping-list-api/internal/item/handler"
	itemRepo "family-shopping-list-api/internal/item/repository"
	itemService "family-shopping-list-api/internal/item/service"
	listHandler "family-shopping-list-api/internal/list/handler"
	listRepo "family-shopping-list-api/internal/list/repository"
	listService "family-shopping-list-api/internal/list/service"
	memberHandler "family-shopping-list-api/internal/member/handler"
	memberRepo "family-shopping-list-api/internal/member/repository"
	memberService "family-shopping-list-api/internal/member/service"
	userHandler "family-shopping-list-api/internal/user/handler"
	userRepo "family-shopping-list-api/internal/user/repository"
	userService "family-shopping-list-api/internal/user/service"
	"family-shopping-list-api/pkg/response"
)

func NewRouter(db *sql.DB, cfg config.Config) http.Handler {
	userRepo := userRepo.New(db)
	listRepo := listRepo.New(db)
	itemRepo := itemRepo.New(db)
	memberRepo := memberRepo.New(db)
	inviteRepo := inviteRepo.New(db)

	userSvc := userService.New(userRepo, cfg)
	listSvc := listService.New(listRepo, memberRepo)
	memberSvc := memberService.New(memberRepo)
	inviteSvc := inviteService.New(inviteRepo, memberRepo)
	itemSvc := itemService.New(itemRepo)

	users := userHandler.New(userSvc)
	lists := listHandler.New(listSvc)
	members := memberHandler.New(memberSvc)
	invites := inviteHandler.New(inviteSvc)
	items := itemHandler.New(itemSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("POST /api/v1/auth/register", users.Register)
	mux.HandleFunc("POST /api/v1/auth/login", users.Login)
	mux.HandleFunc("GET /api/v1/me", users.Me)

	mux.HandleFunc("POST /api/v1/lists", lists.Create)
	mux.HandleFunc("GET /api/v1/lists", lists.List)
	mux.HandleFunc("GET /api/v1/lists/{list_id}", middleware.RequireListMember(memberRepo, false, lists.GetByID))
	mux.HandleFunc("PATCH /api/v1/lists/{list_id}", middleware.RequireListMember(memberRepo, true, lists.Update))
	mux.HandleFunc("DELETE /api/v1/lists/{list_id}", middleware.RequireListMember(memberRepo, true, lists.Delete))

	mux.HandleFunc("GET /api/v1/lists/{list_id}/members", middleware.RequireListMember(memberRepo, false, members.List))
	mux.HandleFunc("DELETE /api/v1/lists/{list_id}/members/{member_id}", middleware.RequireListMember(memberRepo, true, members.Remove))

	mux.HandleFunc("GET /api/v1/lists/{list_id}/invites", middleware.RequireListMember(memberRepo, false, invites.List))
	mux.HandleFunc("POST /api/v1/lists/{list_id}/invites", middleware.RequireListMember(memberRepo, true, invites.Create))
	mux.HandleFunc("POST /api/v1/invites/join", invites.Join)

	mux.HandleFunc("GET /api/v1/lists/{list_id}/items", middleware.RequireListMember(memberRepo, false, items.List))
	mux.HandleFunc("POST /api/v1/lists/{list_id}/items", middleware.RequireListMember(memberRepo, false, items.Create))
	mux.HandleFunc("GET /api/v1/lists/{list_id}/items/{item_id}", middleware.RequireListMember(memberRepo, false, items.Get))
	mux.HandleFunc("PATCH /api/v1/lists/{list_id}/items/{item_id}", middleware.RequireListMember(memberRepo, false, items.Update))
	mux.HandleFunc("DELETE /api/v1/lists/{list_id}/items/{item_id}", middleware.RequireListMember(memberRepo, false, items.Delete))

	return middleware.Auth(cfg.JWTSecret, mux)
}

func health(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]string{"status": "ok"})
}
