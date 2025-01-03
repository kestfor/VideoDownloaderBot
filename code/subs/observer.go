package subs

type Observer interface {
	Update(event any)
}
