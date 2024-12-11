package main

import (
	"fmt"
	"time"

	application "main.go/src/App"
	"main.go/src/controller"
	ui "main.go/src/scenes"
)

func main() {
	fmt.Println("main")
	numVehicles := 100
	timeout := 10 * time.Second
	parkingLotService := application.NewParkingLotService(20, timeout)
	parkingUI := ui.NewParkingUI(parkingLotService)
	parkingLotService.RegisterObserver(parkingUI)

	if parkingLotService == nil {
		fmt.Println("Error: parkingService es nil")
		return
	}

	go controller.StartParkingControl(parkingLotService, numVehicles)
	ui.StartUI(parkingLotService)
}
