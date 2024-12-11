package models

type Observer interface {
	Update(info UpdateInfo)
}

type Subject interface {
	Register(Observer)
	Remove(Observer)
	NotifyAll()
}
