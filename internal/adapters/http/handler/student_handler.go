package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	"github.com/google/uuid"
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
	group := router.Group("/students")

	group.POST("", h.Create)

	protected := group.Group("")
	if auther != nil {
		protected.Use(auther)
	}
	protected.GET("", h.List)
	protected.GET("/:id", h.GetByID)
	protected.DELETE("/:id", h.Anonymize)
}

func (h *StudentHandler) Create(c *gin.Context) {
	var input student.CreateStudentInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos: " + err.Error()})
		return
	}

	out, err := h.createUC.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *StudentHandler) GetByID(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}

	uidParam := c.Param("id")
	studID, err := uuid.Parse(uidParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido, formato uid esperado"})
		return
	}

	out, err := h.getUC.Execute(c.Request.Context(), studID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if claims.Role == "STUDENT" && claims.UserID.String() != out.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "você não tem permissão para visualizar o CPF/Laudo Biométrico de outros perfis"})
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

	uidParam := c.Param("id")
	studID, err := uuid.Parse(uidParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if claims.Role == "STUDENT" {
		out, _ := h.getUC.Execute(c.Request.Context(), studID)
		if out != nil && claims.UserID.String() != out.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "não é permitido solicitar anonimização de terceiros"})
			return
		}
	}

	err = h.anonymizeUC.Execute(c.Request.Context(), studID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao purgar dados: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dados identificáveis removidos com sucesso. Perfil Inativado."})
}

func (h *StudentHandler) List(c *gin.Context) {
	_, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "acesso negado, sessão inválida"})
		return
	}

	out, err := h.listUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}
