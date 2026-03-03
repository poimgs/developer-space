package handler

import (
	"github.com/go-chi/chi/v5"

	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/service"
)

func RegisterRoutes(r chi.Router, memberHandler *MemberHandler, authHandler *AuthHandler, sessionHandler *SessionHandler, rsvpHandler *RSVPHandler, authSvc *service.AuthService, memberRepo middleware.MemberLookup) {
	// Public auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/magic-link", authHandler.RequestMagicLink)
		r.Get("/verify", authHandler.Verify)

		// Authenticated auth routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authSvc, memberRepo))
			r.Post("/logout", authHandler.Logout)
			r.Get("/me", authHandler.Me)
			r.Patch("/profile", authHandler.UpdateProfile)
		})
	})

	// Admin-only member routes
	r.Route("/api/members", func(r chi.Router) {
		r.Use(middleware.Auth(authSvc, memberRepo))
		r.Use(middleware.Admin)
		r.Post("/", memberHandler.Create)
		r.Get("/", memberHandler.List)
		r.Get("/{id}", memberHandler.GetByID)
		r.Patch("/{id}", memberHandler.Update)
		r.Delete("/{id}", memberHandler.Delete)
	})

	// Session routes — all authenticated, admin for create/update/cancel
	r.Route("/api/sessions", func(r chi.Router) {
		r.Use(middleware.Auth(authSvc, memberRepo))
		r.Get("/", sessionHandler.List)
		r.Get("/{id}", sessionHandler.GetByID)

		// RSVP routes — all authenticated
		r.Post("/{id}/rsvp", rsvpHandler.RSVP)
		r.Delete("/{id}/rsvp", rsvpHandler.CancelRSVP)
		r.Get("/{id}/rsvps", rsvpHandler.ListRSVPs)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Admin)
			r.Post("/", sessionHandler.Create)
			r.Patch("/{id}", sessionHandler.Update)
			r.Delete("/{id}", sessionHandler.Cancel)
			r.Delete("/series/{id}", sessionHandler.StopSeries)
		})
	})
}
