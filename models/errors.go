package models

import "errors"

// Errors как переменные (для возврата)
var (
	ErrSelfParent         = errors.New("department cannot be parent of itself")
	ErrCycleDetected      = errors.New("cannot move department to its own child (cycle detected)")
	ErrDepartmentNotFound = errors.New("department not found")
	ErrTargetNotFound     = errors.New("target department not found")
	ErrNameEmpty          = errors.New("department name cannot be empty")
	ErrNameTooLong        = errors.New("department name too long (max 200)")
	ErrParentNotFound     = errors.New("parent department not found")
	ErrNameExists         = errors.New("department with this name already exists in this parent")
	ErrInvalidMode        = errors.New("invalid mode, use 'cascade' or 'reassign'")
	ErrReassignToSame     = errors.New("cannot reassign to the same department")
)

// Для Employee
var (
	ErrFullNameEmpty   = errors.New("full name cannot be empty")
	ErrFullNameTooLong = errors.New("full name too long (max 200 characters)")
	ErrPositionEmpty   = errors.New("position cannot be empty")
	ErrPositionTooLong = errors.New("position too long (max 200 characters)")
	ErrHiredAtFuture   = errors.New("hired_at cannot be in the future")
)
