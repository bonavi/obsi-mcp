package errors

import "pkg/errors"

// Is проверяет цепочку ошибок на совпадение с целевой
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As приводит ошибку к целевому типу
func As(get error, target any) bool {
	return errors.As(get, target)
}

// CastError приводит ошибку к кастомной. Для nil возвращает пустую ошибку без типа
func CastError(err error) errors.Error {
	if err == nil {
		return errors.Error{}
	}
	return errors.CastError(err)
}
