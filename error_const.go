package start_hw31

import "errors"

var (
	ErrUserAlreadyExist      = errors.New("User with that username and password already exist")
	ErrUsernamePasswordEmpty = errors.New("Field username or password is empty")
)
