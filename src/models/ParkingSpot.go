package models

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/oakmound/oak/v4/alg/floatgeom"
)

type ParkingSpot struct {
	rect         *floatgeom.Rect2
	occupied     bool
	carImage     *canvas.Image
	carContainer *fyne.Container
	position     fyne.Position
	spotLabel    *canvas.Text
	carID        int
}

func (spot *ParkingSpot) GetRect() *floatgeom.Rect2 {
	return spot.rect
}
