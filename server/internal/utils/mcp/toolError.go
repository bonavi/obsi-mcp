package mcp

import (
	stdErrors "errors"
	"fmt"
	pkgErrors "pkg/errors"
	"pkg/log"
	"server/internal/utils/errors"
)

// ToolError превращает ошибку сервиса в текст для агента: человеческий текст типа и причина.
// В отличие от Tasks детали внутренних ошибок (вывод git, пути) остаются только в журнале
func ToolError(err error) error {
	log.LogError(err)
	castError := pkgErrors.CastError(err)
	if castError.ErrorType == errors.InternalServer || castError.ErrorType == pkgErrors.Default {
		return stdErrors.New(errors.InternalServer.HumanText)
	}
	return fmt.Errorf("%s: %s", castError.ErrorType.HumanText, castError.Err.Error())
}
