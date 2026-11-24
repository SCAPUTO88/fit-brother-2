package errs

import "errors"

var (
	ErrInvalidInput = errors.New("dados inválidos")
	ErrUnauthorized = errors.New("não autenticado")
	ErrForbidden    = errors.New("acesso negado")
	ErrNotFound     = errors.New("recurso não encontrado")
	ErrConflict     = errors.New("conflito de dados")
)
