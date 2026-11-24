package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type StudentHandler struct {
	createUC    *student.CreateStudentUseCase
	getUC       *student.GetStudentUseCase
	listUC      *student.ListStudentsUseCase
	anonymizeUC *student.AnonymizeStudentUseCase
}

func NewStudentHandler(
	createUC *student.CreateStudentUseCase,
	getUC *student.GetStudentUseCase,
	listUC *student.ListStudentsUseCase,
	anonymizeUC *student.AnonymizeStudentUseCase,
) *StudentHandler {
	return &StudentHandler{
		createUC:    createUC,
		getUC:       getUC,
		listUC:      listUC,
		anonymizeUC: anonymizeUC,
	}
}

func (h *StudentHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	protected := router.Group("/students")
	if auther != nil {
		protected.Use(auther)
	}
	protected.POST("", h.Create)
	protected.GET("", h.List)
	protected.GET("/:id", h.GetByID)
	protected.DELETE("/:id", h.Anonymize)
}

func (h *StudentHandler) Create(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem cadastrar alunos"})
		return
	}

	var input student.CreateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}
	input.TrainerID = claims.UserID

	out, err := h.createUC.Execute(c.Request.Context(), input)
	if err != nil {
		respondDomainError(c, err, "falha ao cadastrar aluno")
		return
	}

	c.JSON(http.StatusCreated, out)
}

func (h *StudentHandler) GetByID(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}

	studID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido, formato uuid esperado"})
		return
	}

	out, err := h.getUC.Execute(c.Request.Context(), studID, claims.UserID, claims.Role)
	if err != nil {
		respondDomainError(c, err, "falha ao consultar aluno")
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *StudentHandler) Anonymize(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado"})
		return
	}

	studID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.anonymizeUC.Execute(c.Request.Context(), studID, claims.UserID, claims.Role); err != nil {
		respondDomainError(c, err, "falha ao anonimizar aluno")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dados identificáveis removidos com sucesso. Perfil Inativado."})
}

func (h *StudentHandler) List(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}
	if claims.Role != user.RoleTrainer {
		c.JSON(http.StatusForbidden, gin.H{"error": "apenas treinadores podem listar alunos"})
		return
	}

	out, err := h.listUC.Execute(c.Request.Context(), claims.UserID)
	if err != nil {
		respondDomainError(c, err, "falha ao listar alunos")
		return
	}

	c.JSON(http.StatusOK, out)
}
