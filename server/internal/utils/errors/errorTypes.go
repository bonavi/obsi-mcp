package errors

import (
	"net/http"
	"pkg/errors"
)

// InternalServer — внутренняя ошибка, детали клиенту не показываем
var InternalServer = errors.ErrorType{
	Name:      "InternalServer",
	HTTPCode:  http.StatusInternalServerError,
	LogAs:     errors.LogAsError,
	HumanText: "internal server error",
}

// BadRequest — некорректные входные данные
var BadRequest = errors.ErrorType{
	Name:      "BadRequest",
	HTTPCode:  http.StatusBadRequest,
	LogAs:     errors.LogAsWarning,
	HumanText: "bad request",
}

// Forbidden — доступ к пути запрещён
var Forbidden = errors.ErrorType{
	Name:      "Forbidden",
	HTTPCode:  http.StatusForbidden,
	LogAs:     errors.LogAsWarning,
	HumanText: "path is not allowed",
}

// NotFound — заметка не найдена
var NotFound = errors.ErrorType{
	Name:      "NotFound",
	HTTPCode:  http.StatusNotFound,
	LogAs:     errors.LogAsWarning,
	HumanText: "note not found",
}

// Conflict — состояние заметки не совпало с ожидаемым
var Conflict = errors.ErrorType{
	Name:      "Conflict",
	HTTPCode:  http.StatusConflict,
	LogAs:     errors.LogAsWarning,
	HumanText: "note state conflict",
}
