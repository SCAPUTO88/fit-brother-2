package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/SCAPUTO88/fit-brother/internal/application/student"
)

type StudentHandler struct {
	createUC *student.CreateStudentUseCase
}

func NewStudentHandler(createUC *student.CreateStudentUseCase) *StudentHandler {
	return &StudentHandler{createUC: createUC}
}

// RegisterRoutes acopla os endpoints do dominio de Student ao Router, adicionando os middlewares de protecao JWT.
// Por segurança, apenas TRAINERS (ou o proprio admin) deveriam conseguir cadastrar alunos.
func (h *StudentHandler) RegisterRoutes(router *gin.RouterGroup) {
	// A rota base "students".
	// O auth_middleware será adicionado lá no router central para todo esse subgrupo, 
	// mas poderiamos colocar um RoleMiddleware("TRAINER") aqui também.
	group := router.Group("/students")
	{
		group.POST("", h.Create)
	}
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
