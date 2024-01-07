package models

import "errors"

var (
	EntityNotFound      = errors.New("Entity Not Found")
	EntityAlreadyExists = errors.New("Entity already exists")
)
