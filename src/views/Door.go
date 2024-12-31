package views

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

const (
	doorWidth  = 100
	doorHeight = 60
)

func CreateGradientDoor(isEntry bool) *fyne.Container {
	doorColor := color.RGBA{0, 204, 0, 255}
	if !isEntry {
		doorColor = color.RGBA{0, 0, 102, 255}
	}

	mainRect := canvas.NewRectangle(doorColor)
	border := canvas.NewRectangle(theme.ShadowColor())
	border.StrokeWidth = 2
	border.StrokeColor = theme.ForegroundColor()

	label := canvas.NewText(map[bool]string{true: "ENTRADA", false: "SALIDA"}[isEntry], theme.ForegroundColor())
	label.TextSize = 12
	label.TextStyle.Bold = true

	door := container.NewWithoutLayout(mainRect, border, label)

	mainRect.Resize(fyne.NewSize(doorWidth, doorHeight))
	border.Resize(fyne.NewSize(doorWidth, doorHeight))
	label.Move(fyne.NewPos(doorWidth/4, doorHeight/3))

	return door
}
