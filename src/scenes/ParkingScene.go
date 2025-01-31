package ui

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	application "main.go/src/App"
	"main.go/src/views"
)

const (
	gridSize      = 20
	cellSize      = 60
	gridRows      = 4
	gridCols      = 5
	doorWidth     = 100
	doorHeight    = 60
	marginTop     = 120
	marginLeft    = 150
	laneWidth     = 40
	cornerRadius  = 5
	animationStep = 50 * time.Millisecond
)

type PathPoint struct {
	x, y float32
}

type ParkingSpot struct {
	rect         *canvas.Rectangle
	occupied     bool
	carImage     *canvas.Image
	carContainer *fyne.Container
	position     fyne.Position
	spotLabel    *canvas.Text
	carID        int
}

type ParkingUI struct {
	spots         [gridSize]ParkingSpot
	counterSpots  int
	entryDoor     *fyne.Container
	container     *fyne.Container
	service       *application.ParkingLotService
	updateChannel chan application.UpdateInfo
	statusLabel   *widget.Label
	app           fyne.App
	window        fyne.Window
	driveLanes    []PathPoint
	background    *canvas.Rectangle
	spotMutex     sync.RWMutex
}

func (ui *ParkingUI) Update(info application.UpdateInfo) {
	ui.updateChannel <- info
}

func StartUI(service *application.ParkingLotService) {
	parkingUI := NewParkingUI(service)
	if parkingUI == nil {
		return
	}

	service.RegisterObserver(parkingUI)

	header := parkingUI.createHeader()
	content := container.NewVBox(
		header,
		parkingUI.container,
	)

	parkingUI.window.SetContent(content)
	go parkingUI.processUpdates()
	parkingUI.window.ShowAndRun()
}

func NewParkingUI(service *application.ParkingLotService) *ParkingUI {
	a := app.New()
	w := a.NewWindow("Estacionamiento")
	w.SetFixedSize(false)
	w.Resize(fyne.NewSize(600, 600))
	ui := &ParkingUI{
		service:       service,
		window:        w,
		container:     container.NewWithoutLayout(),
		driveLanes:    make([]PathPoint, 0),
		updateChannel: make(chan application.UpdateInfo, 100),
		app:           a,
		counterSpots:  0,
	}
	ui.window.CenterOnScreen()

	ui.initializeUI()
	return ui
}

func (ui *ParkingUI) initializeUI() {
	ui.background = canvas.NewRectangle(theme.BackgroundColor())
	ui.container.Add(ui.background)

	ui.createDriveLanes()
	ui.setupParkingSpots()
	ui.drawDriveLanes()

	ui.entryDoor = views.CreateGradientDoor(true)
	entryPos := fyne.NewPos(float32(marginLeft/2), float32(marginTop+cellSize*2))
	ui.entryDoor.Move(entryPos)
	ui.container.Add(ui.entryDoor)
}

func (ui *ParkingUI) setupParkingSpots() {
	for i := 0; i < gridSize; i++ {
		row := i / gridCols
		col := i % gridCols

		spot := &ui.spots[i]
		spot.rect = canvas.NewRectangle(color.RGBA{255, 255, 0, 255})
		spot.rect.StrokeWidth = 1
		spot.rect.StrokeColor = theme.PrimaryColor()
		spot.rect.Resize(fyne.NewSize(cellSize-3, cellSize-3))

		spot.position = fyne.NewPos(
			float32(marginLeft+col*cellSize),
			float32(marginTop+row*cellSize),
		)
		spot.rect.Move(spot.position)

		spot.spotLabel = canvas.NewText(fmt.Sprintf("%d", i+1), theme.ForegroundColor())
		spot.spotLabel.TextSize = 12
		spot.spotLabel.Move(fyne.NewPos(
			spot.position.X+5,
			spot.position.Y+5,
		))

		ui.container.Add(spot.rect)
		ui.container.Add(spot.spotLabel)
	}
}

func (ui *ParkingUI) createHeader() *fyne.Container {
	headerLabel := widget.NewLabel("Sistema de Estacionamiento")
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}
	headerLabel.Alignment = fyne.TextAlignCenter

	ui.statusLabel = widget.NewLabel("Estado: 0/20 espacios ocupados")
	ui.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	headerContainer := container.NewVBox(
		headerLabel,
		ui.statusLabel,
	)

	return headerContainer
}

func (ui *ParkingUI) createDriveLanes() {
	startX := marginLeft/2 + doorWidth
	endX := marginLeft + cellSize*gridCols
	laneY := marginTop + cellSize*2 + doorHeight/2
	for x := startX; x <= endX; x += 10 {
		ui.driveLanes = append(ui.driveLanes, PathPoint{float32(x), float32(laneY)})
	}
	for col := 0; col < gridCols; col++ {
		x := marginLeft + col*cellSize + cellSize/2
		for y := marginTop; y <= marginTop+cellSize*gridRows; y += 10 {
			ui.driveLanes = append(ui.driveLanes, PathPoint{float32(x), float32(y)})
		}
	}
}

func (ui *ParkingUI) drawDriveLanes() {
	laneColor := theme.DisabledButtonColor()
	mainLane := canvas.NewRectangle(laneColor)
	mainLane.Resize(fyne.NewSize(float32(0.5*gridCols*cellSize), laneWidth))
	mainLane.Move(fyne.NewPos(
		float32(marginLeft/2+doorWidth/2),
		float32(marginTop+cellSize*2+doorHeight/2-laneWidth/2),
	))
	mainLane.StrokeWidth = 5
	mainLane.StrokeColor = color.RGBA{255, 255, 255, 255}
	ui.container.Add(mainLane)
	for col := 0; col < gridCols; col++ {
		vertLane := canvas.NewRectangle(laneColor)
		vertLane.Resize(fyne.NewSize(laneWidth, float32(gridRows*cellSize)))
		vertLane.Move(fyne.NewPos(
			float32(marginLeft+col*cellSize+cellSize/2-laneWidth/2),
			float32(marginTop),
		))
		ui.container.Add(vertLane)
	}
}

func (ui *ParkingUI) processUpdates() {
	for update := range ui.updateChannel {
		fmt.Println("Hay un update")
		ui.safeUpdate(update)
	}
}

func (ui *ParkingUI) safeUpdate(info application.UpdateInfo) {
	defer ui.spotMutex.Unlock()
	ui.spotMutex.Lock()

	switch info.EventType {
	case "CarParked":
		ui.parkCar(info)
	case "CarExiting":
		ui.removeCar(info)
	default:
		ui.updateStatusLabel()
	}
}

func (ui *ParkingUI) parkCar(info application.UpdateInfo) {
	spot := ui.findAvailableSpot()
	if spot == nil {
		fmt.Println("No hay spots disponibles")
		return
	}

	spot.carID = info.Car.GetId()
	carContainer, carImage, _ := views.CreateCarWithContainer(info.Car.GetImage())
	spot.carContainer = carContainer
	spot.carImage = carImage

	spot.rect.FillColor = color.RGBA64{0, 204, 0, 255}
	spot.occupied = true

	carContainer.Move(spot.position)
	ui.container.Add(carContainer)
	fmt.Println("[Added-Before] Spot: ", ui.counterSpots)
	ui.counterSpots++
	fmt.Println("[Added-After] Spot: ", ui.counterSpots)

	ui.updateStatusLabel()
}

func (ui *ParkingUI) removeCar(info application.UpdateInfo) {
	for i := range ui.spots {
		spot := &ui.spots[i]
		if spot.occupied && spot.carID == info.Car.GetId() {
			spot.rect.FillColor = color.RGBA{255, 255, 0, 255}
			spot.occupied = false
			if spot.carContainer != nil {
				spot.carContainer.Hide()
				ui.container.Remove(spot.carContainer)
				spot.carContainer = nil
				fmt.Println("[Removed-Before] Spot: ", ui.counterSpots)
				ui.counterSpots--
				fmt.Println("[Removed-After] Spot: ", ui.counterSpots)
			}
			break
		}
	}
	ui.updateStatusLabel()
}

func (ui *ParkingUI) updateStatusLabel() {
	fmt.Println("[LABEL] Actualizando estado")
	ui.statusLabel.SetText(fmt.Sprintf("Estado: %d/%d espacios ocupados", ui.counterSpots, gridSize))
	currentSize := ui.window.Canvas().Size()
	ui.window.Resize(fyne.NewSize(currentSize.Width+1, currentSize.Height))
	ui.window.Resize(currentSize)
	fmt.Println("[LABEL] Estado actualizado")
}

func (ui *ParkingUI) findAvailableSpot() *ParkingSpot {
	for i := range ui.spots {
		if !ui.spots[i].occupied {
			return &ui.spots[i]
		}
	}
	return nil
}
