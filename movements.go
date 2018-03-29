package termination

func UpMovement(t *Termination, e *Entity, position Position) Position {
	position.Y -= 1
	return position
}

func DownMovement(t *Termination, e *Entity, position Position) Position {
	position.Y += 1
	return position
}

func LeftMovement(t *Termination, e *Entity, position Position) Position {
	position.X -= 1
	return position
}

func RightMovement(t *Termination, e *Entity, position Position) Position {
	position.X += 1
	return position
}
