package models

import (
	"math/rand"
	"time"
)

type Car struct {
	ID        int
	State     string
	Spot      int
	Image     string
	Cancel    chan bool
	observers []Observer
	x, y      float64
}

func NewCar(id int, image string) *Car {
	modelPaths := []string{
		"assets/img/carroCafe.png",
		"assets/img/carroNaranja.png",
		"assets/img/carroRojo.png",
		"assets/img/carroVerde.png",
	}
	rand := (rand.NewSource(time.Now().UnixNano()))
	image = modelPaths[rand.Int63()]
	return &Car{
		ID:        id,
		Image:     image,
		State:     "Waiting",
		Cancel:    make(chan bool, 1),
		observers: []Observer{},
	}
}

func (c *Car) GetId() int {
	return c.ID
}

func (c *Car) GetImage() string {
	return c.Image
}

func (c *Car) Register(observer Observer) {
	c.observers = append(c.observers, observer)
}
