package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func CreateCarWithContainer(filePath string) (*fyne.Container, *canvas.Image, error) {
	carImg := canvas.NewImageFromFile(filePath)
	carImg.FillMode = canvas.ImageFillOriginal
	carImg.Resize(fyne.NewSize(60, 60))

	carContainer := container.NewWithoutLayout(carImg)
	carContainer.Resize(fyne.NewSize(5, 5))

	return carContainer, carImg, nil
}
