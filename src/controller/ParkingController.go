package controller

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	application "main.go/src/App"
	"main.go/src/models"
)

func StartParkingControl(service *application.ParkingLotService, numVehicles int) {
	var wg sync.WaitGroup
	wg.Add(numVehicles)

	entryChannel := service.GetEntryChannel()
	exitChannel := service.GetExitChannel()
	modelPaths := []string{
		"/Users/Herobrot/Documents/Cuatri-7B/ProgramacionConcurrente/Corte_2/estacionamiento/Assets/carroNaranja.png",
		"/Users/Herobrot/Documents/Cuatri-7B/ProgramacionConcurrente/Corte_2/estacionamiento/Assets/carroCafe.png",
		"/Users/Herobrot/Documents/Cuatri-7B/ProgramacionConcurrente/Corte_2/estacionamiento/Assets/carroRojo.png",
		"/Users/Herobrot/Documents/Cuatri-7B/ProgramacionConcurrente/Corte_2/estacionamiento/Assets/carroVerde.png",
	}

	go func() {
		for i := 0; i < numVehicles; i++ {
			rand.Seed(time.Now().UnixNano())
			image := modelPaths[rand.Intn(len(modelPaths))]
			car := models.Car{
				ID:     i,
				State:  "Waiting",
				Image:  image,
				Spot:   i,
				Cancel: make(chan bool, 1),
			}

			fmt.Printf("Generando vehículo %d\n", car.ID)
			time.Sleep(time.Duration(rand.Intn(300)+3) * time.Millisecond)
			service.EnterParking(&car)
			fmt.Printf("Vehiculo %d Entrando al lugar del parking.\n", car.ID)
			wg.Add(1)
		}

		close(entryChannel)
		fmt.Println("Canal de entrada cerrado.")
	}()

	go func() {
		for vehicle := range exitChannel {
			go func(v *models.Car) {
				service.ExitParking(v)
				time.Sleep(time.Duration(rand.Intn(3)+3) * time.Second)
				fmt.Printf("Vehicle %d Saliendo del  parking.\n", v.ID)
				wg.Done()
			}(vehicle)
		}
	}()

	wg.Wait()
	fmt.Println("Todos los  vehicles han  salido .")
}

func GenerateVehicles(entryChannel chan<- *models.Car, numVehicles int) {
	id := 1
	for i := 0; i < numVehicles; i++ {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		vehicle := &models.Car{State: "Waiting"}
		fmt.Printf("Generando vehículo %d\n", id)
		entryChannel <- vehicle
		fmt.Printf("Vehículo %d enviado al canal\n", id)
		id++
	}

	close(entryChannel)
	fmt.Println("Todos los vehículos han sido generados.")
}
