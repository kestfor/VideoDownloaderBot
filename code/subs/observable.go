package subs

type Observable interface {
	NotifyObservers(event any)
	AddObserver(observer Observer) error
	DetachObserver(observer Observer) error
}
