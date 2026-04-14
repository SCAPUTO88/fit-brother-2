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
	anonymizeUC *student.AnonymizeStudentUseCase
}

func NewStudentHandler(
	createUC *student.CreateStudentUseCase,
	getUC *student.GetStudentUseCase,
	anonymizeUC *student.AnonymizeStudentUseCase,
) *StudentHandler {
	return &StudentHandler{
		createUC:    createUC,
		getUC:       getUC,
		anonymizeUC: anonymizeUC,
	}
}

// RegisterRoutes acopla os endpoints do dominio de Student ao Router, adicionando os middlewares de protecao JWT.
// Por segurança, apenas TRAINERS (ou o proprio admin) deveriam conseguir cadastrar alunos.
func (h *StudentHandler) RegisterRoutes(router *gin.RouterGroup, auther gin.HandlerFunc) {
	group := router.Group("/students")
	
	// Pública
	group.POST("", h.Create)

	// Rotas em que DEVE haver Token Validado passando antes via RouterGroup (auth_middleware)
	// Para leitura profunda de ficha (Com revelação do CPF) e LGPD Anonymizer
	protected := group.Group("")
	if auther != nil {
		protected.Use(auther)
	}
	protected.GET("/:id", h.GetByID)
	protected.DELETE("/:id", h.Anonymize)
}

func (h *StudentHandler) Create(c *gin.Context) {
	var input student.CreateStudentInput
	
	// BindJSON fará a checagem das annotations de "binding:required,email,etc"
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

	// === LGPD Authorization Gate === //
	// Impede que Alunos consigam pegar laudos e CPFs de outros assinantes
	if claims.Role == "STUDENT" && claims.UserID.String() != out.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "você não tem persmissão para visualizar o CPF/Laudo Biométrico de outros perfis"})
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

	// === Segurança de Direito ao Esquecimento === //
	// Alunos solicitando esquecimento só podem pedir do seu proprio perfil
	if claims.Role == "STUDENT" {
		out, _ := h.getUC.Execute(c.Request.Context(), studID)
		if out != nil && claims.UserID.String() != out.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "LGPD Breach Act: não é permitido solicitar anonimização de terceiros"})
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
