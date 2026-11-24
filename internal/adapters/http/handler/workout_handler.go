package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	appworkout "github.com/SCAPUTO88/fit-brother/internal/application/workout"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type WorkoutHandler struct {
	createPlanUC     *appworkout.CreatePlanUseCase
	createSessionUC  *appworkout.CreateSessionUseCase
	createExerciseUC *appworkout.CreateExerciseUseCase
	listPlansUC      *appworkout.ListPlansUseCase
	getPlanDetailsUC *appworkout.GetPlanDetailsUseCase
}

func NewWorkoutHandler(
	cp *appworkout.CreatePlanUseCase,
	cs *appworkout.CreateSessionUseCase,
	ce *appworkout.CreateExerciseUseCase,
	lp *appworkout.ListPlansUseCase,
	gpd *appworkout.GetPlanDetailsUseCase,
) *WorkoutHandler {
	return &WorkoutHandler{
		createPlanUC:     cp,
		createSessionUC:  cs,
		createExerciseUC: ce,
		listPlansUC:      lp,
		getPlanDetailsUC: gpd,
	}
}

func (h *WorkoutHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	g := router.Group("/workouts")
	if auther != nil {
		g.Use(auther)
	}

	g.POST("", h.CreatePlan)
	g.GET("", h.ListPlans)
	g.GET("/:id", h.GetPlanDetails)
	g.POST("/:id/sessions", h.CreateSession)
	g.POST("/sessions/:session_id/exercises", h.CreateExercise)
}

func (h *WorkoutHandler) CreatePlan(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem criar planos de treino"})
		return
	}

	var input appworkout.CreatePlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	input.TrainerID = claims.UserID.String()

	out, err := h.createPlanUC.Execute(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar plano de treino")
		return
	}

	c.JSON(http.StatusCreated, out)
}

func (h *WorkoutHandler) GetPlanDetails(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	planIDStr := c.Param("id")

	out, err := h.getPlanDetailsUC.Execute(c.Request.Context(), planIDStr, claims.UserID, claims.Role)
	if err != nil {
		respondDomainError(c, err, "falha ao carregar detalhes do plano")
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *WorkoutHandler) ListPlans(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	var plans []*appworkout.PlanOutput
	var err error

	switch claims.Role {
	case user.RoleTrainer:
		plans, err = h.listPlansUC.ByTrainer(c.Request.Context(), claims.UserID)
	case user.RoleStudent:
		plans, err = h.listPlansUC.ByStudent(c.Request.Context(), claims.UserID)
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "role desconhecida"})
		return
	}

	if err != nil {
		respondDomainError(c, err, "falha ao listar planos de treino")
		return
	}

	c.JSON(http.StatusOK, plans)
}

func (h *WorkoutHandler) CreateSession(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem adicionar sessões"})
		return
	}

	planIDStr := c.Param("id")
	if _, err := uuid.Parse(planIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do plano inválido"})
		return
	}

	var input appworkout.CreateSessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	input.WorkoutPlanID = planIDStr

	out, err := h.createSessionUC.Execute(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar sessão")
		return
	}

	c.JSON(http.StatusCreated, out)
}

func (h *WorkoutHandler) CreateExercise(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem adicionar exercícios"})
		return
	}

	sessionIDStr := c.Param("session_id")
	if _, err := uuid.Parse(sessionIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da sessão inválido"})
		return
	}

	var input appworkout.CreateExerciseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	input.WorkoutSessionID = sessionIDStr

	out, err := h.createExerciseUC.Execute(c.Request.Context(), claims.UserID, input)
	if err != nil {
		respondDomainError(c, err, "falha ao criar exercício")
		return
	}

	c.JSON(http.StatusCreated, out)
}
