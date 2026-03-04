package handler

import (
	"net/http"

	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

type SkillsHandler struct {
	svc *service.MemberService
}

func NewSkillsHandler(svc *service.MemberService) *SkillsHandler {
	return &SkillsHandler{svc: svc}
}

func (h *SkillsHandler) List(w http.ResponseWriter, r *http.Request) {
	skills, err := h.svc.ListDistinctSkills(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, skills)
}
