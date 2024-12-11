package models

import (
	"fmt"
	"math/rand"
	"simulador/src/models"
	"sync"
	"time"
)

type UpdateInfo struct {
	Car       *models.Car
	Entering  bool
	Spot      int
	EventType string
}

type ParkingLotService struct {
	ParkingLot      *models.ParkingLot
	entryChannel    chan *models.Car
	exitChannel     chan *models.Car
	entryQueue      []*models.Car
	spotAvailable   *sync.Cond
	queueMutex      sync.Mutex
	UpdateChannel   chan UpdateInfo
	timeoutDuration time.Duration
	observers       []Observer
	observerMutex   sync.Mutex
}

type ParkingLot struct {
	Capacity int
	Vehicles int
	Spots    []int
	Mutex    sync.Mutex
}

func StartParkingControl(parking *ParkingLotService, numVehicles int) {
	var wg sync.WaitGroup
	wg.Add(numVehicles)

	entryChannel := parking.GetEntryChannel()
	exitChannel := parking.GetExitChannel()

	go func() {
		for i := 0; i < numVehicles; i++ {

			car := models.Car{
				ID:     i,
				State:  "Waiting",
				Spot:   i,
				Cancel: make(chan bool, 1),
			}

			fmt.Printf("Generando vehículo %d\n", car.ID)
			time.Sleep(time.Duration(rand.Intn(3)+3) * time.Second)
			parking.EnterParking(&car)
			fmt.Printf("Vehiculo %d Entrando al lugar del parking.\n", car.ID)
			wg.Add(1)
		}

		close(entryChannel)
		fmt.Println("Canal de entrada cerrado.")
	}()

	go func() {
		for vehicle := range exitChannel {
			go func(v *models.Car) {
				parking.ExitParking(v)
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
	for i := 0; i < numVehicles; i++ {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		vehicle := &models.Car{State: "Waiting"}
		fmt.Printf("Generando vehículo %d\n", i)
		entryChannel <- vehicle
		fmt.Printf("Vehículo %d enviado al canal\n", i)
	}

	close(entryChannel)
	fmt.Println("Todos los vehículos han sido generados.")
}

func NewParkingLotService(capacity int, timeout time.Duration) *ParkingLotService {
	spots := make([]int, capacity)
	parkingLot := models.ParkingLot{Capacity: capacity, Spots: spots}
	ps := &ParkingLotService{
		ParkingLot:      parkingLot,
		entryChannel:    make(chan *models.Car),
		exitChannel:     make(chan *models.Car),
		entryQueue:      make([]*models.Car, 20),
		timeoutDuration: timeout,
		UpdateChannel:   make(chan UpdateInfo),
	}
	ps.spotAvailable = sync.NewCond(&sync.Mutex{})
	return ps
}

func (ps *ParkingLotService) RegisterObserver(o Observer) {
	ps.observerMutex.Lock()
	defer ps.observerMutex.Unlock()
	ps.observers = append(ps.observers, o)
}

func (ps *ParkingLotService) RemoveObserver(o Observer) {
	ps.observerMutex.Lock()
	defer ps.observerMutex.Unlock()
	for i, observer := range ps.observers {
		if observer == o {
			ps.observers = append(ps.observers[:i], ps.observers[i+1:]...)
			break
		}
	}
}

func (ps *ParkingLotService) notifyObservers(info UpdateInfo) {
	defer ps.observerMutex.Unlock()
	fmt.Printf("Notificando: Car %d - Estado: %s, Spot: %d, Tipo de Evento: %s\n", info.Car.ID, info.Spot, info.Car.Spot, info.EventType)
	ps.observerMutex.Lock()
	for _, observer := range ps.observers {
		go observer.Update(info)
	}
}

func (ps *ParkingLotService) EnterParking(car *models.Car) {
	defer ps.spotAvailable.L.Unlock()
	ps.spotAvailable.L.Lock()
	fmt.Printf("Car %d intentando  entrar\n", car.ID)

	if ps.ParkingLot.IsFull() {
		ps.queueMutex.Lock()
		ps.entryQueue = append(ps.entryQueue, car)
		ps.queueMutex.Unlock()
		fmt.Printf("Parking lleno. Carro %d Esperando.\n", car.ID)
		ps.notifyObservers(UpdateInfo{Car: car, Entering: true, Spot: -1, EventType: "CarWaiting"})
		go ps.handleEntryTimeout(car)
		ps.waitForSpot(car)
	} else {
		ps.assignSpotAndEnter(car)
	}
}

func (ps *ParkingLotService) assignSpotAndEnter(car *models.Car) {
	spot := ps.ParkingLot.FindAvailableSpot()
	if spot == -1 {
		fmt.Printf("Error: No hay lugar que se pueda asignar  %d\n", car.ID)
		return
	}
	ps.ParkingLot.ParkCar(car, spot)
	car.State = "Parked"
	fmt.Printf("Carro %d Parkeado en el lugar %d\n", car.ID, spot)
	ps.notifyObservers(UpdateInfo{Car: car, Entering: true, Spot: spot, EventType: "CarParked"})
	go ps.scheduleExit(car)
}

func (ps *ParkingLotService) ExitParking(car *models.Car) {
	ps.spotAvailable.L.Lock()
	defer ps.spotAvailable.L.Unlock()

	if !ps.ParkingLot.IsCarParked(car) {
		fmt.Printf("Vehicle %d no puede salir: Lugar Invalido.\n", car.ID)
		return
	}

	ps.ParkingLot.RemoveCar(car)
	car.State = "Exiting"
	ps.notifyObservers(UpdateInfo{Car: car, Entering: false, Spot: -1, EventType: "CarExiting"})
	ps.spotAvailable.Broadcast()
}

func (ps *ParkingLotService) GetEntryChannel() chan<- *models.Car {
	fmt.Printf("Carro %d entrando al  parking \n")
	return ps.entryChannel
}

func (ps *ParkingLotService) GetExitChannel() chan *models.Car {
	return ps.exitChannel
}

func (ps *ParkingLotService) scheduleExit(car *models.Car) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	ps.exitChannel <- car
}

func (ps *ParkingLotService) waitForSpot(car *models.Car) {
	fmt.Printf("Carr %d Esperando por un lugar \n", car.ID)
	for {
		ps.spotAvailable.L.Lock()
		if !ps.isCarInQueue(car) {
			ps.spotAvailable.L.Unlock()
			return
		}
		if !ps.ParkingLot.IsFull() {
			ps.assignSpotAndEnter(car)
			ps.spotAvailable.L.Unlock()
			return
		}
		ps.spotAvailable.Wait()
		ps.spotAvailable.L.Unlock()
	}
}

func (ps *ParkingLotService) handleEntryTimeout(car *models.Car) {
	timer := time.NewTimer(ps.timeoutDuration)
	select {
	case <-timer.C:
		ps.spotAvailable.L.Lock()
		ps.removeFromQueue(car)
		car.State = "Timeout"
		ps.notifyObservers(UpdateInfo{Car: car, Entering: false, Spot: -1, EventType: "CarTimeout"})
		ps.spotAvailable.L.Unlock()
		fmt.Printf("Car %d se acabo el tiempo de espera %v. saliendo.\n", car.ID, ps.timeoutDuration)

	case <-car.Cancel:
		timer.Stop()
		ps.spotAvailable.L.Lock()
		ps.removeFromQueue(car)
		car.State = "Cancelado"
		ps.notifyObservers(UpdateInfo{Car: car, Entering: false, Spot: -1, EventType: "CarCancelled"})
		ps.spotAvailable.L.Unlock()
		fmt.Printf("Car %d se cancela la espera. Saliendo.\n", car.ID)
	}
}

func (ps *ParkingLotService) removeFromQueue(car *models.Car) {
	defer ps.queueMutex.Unlock()
	ps.queueMutex.Lock()
	for i, c := range ps.entryQueue {
		if c.ID == car.ID {
			ps.entryQueue = append(ps.entryQueue[:i], ps.entryQueue[i+1:]...)
			return
		}
	}
}

func (ps *ParkingLotService) isCarInQueue(car *models.Car) bool {
	defer ps.queueMutex.Unlock()
	ps.queueMutex.Lock()
	for _, c := range ps.entryQueue {
		if c.ID == car.ID {
			return true
		}
	}
	return false
}

func NewParkingService(parkingSize int) *ParkingLotService {
	ps := &ParkingLotService{
		entryChannel: make(chan *models.Car, parkingSize),
		exitChannel:  make(chan *models.Car, parkingSize),
	}
	go ps.handleCarEntry()
	go ps.handleCarExit()
	return ps
}

func (ps *ParkingLotService) handleCarEntry() {
	for car := range ps.entryChannel {
		ps.EnterParking(car)
	}

}

func (ps *ParkingLotService) handleCarExit() {
	for car := range ps.exitChannel {
		ps.ExitParking(car)
	}
}

func NewParkingLot(capacity int) *ParkingLot {
	return &ParkingLot{
		Capacity: capacity,
		Spots:    make([]int, capacity),
	}
}

func (p *ParkingLot) ParkCar(car *Car, spot int) {
	defer p.Mutex.Unlock()
	p.Mutex.Lock()
	p.Vehicles++
	p.Spots[spot] = car.ID
	car.Spot = spot
}

func (p *ParkingLot) RemoveCar(car *Car) {
	defer p.Mutex.Unlock()
	p.Mutex.Lock()
	if car.Spot != -1 {
		p.Vehicles--
		p.Spots[car.Spot] = 0
		car.Spot = -1
	}
}

func (p *ParkingLot) IsCarParked(car *Car) bool {
	defer p.Mutex.Unlock()
	p.Mutex.Lock()
	return car.Spot != -1 && p.Spots[car.Spot] == car.ID
}

func (p *ParkingLot) FindAvailableSpot() int {
	defer p.Mutex.Unlock()
	p.Mutex.Lock()
	for i := 0; i < p.Capacity; i++ {
		if p.Spots[i] == 0 {
			return i
		}
	}
	return -1
}

func (p *ParkingLot) IsFull() bool {
	defer p.Mutex.Unlock()
	p.Mutex.Lock()
	return p.Vehicles >= p.Capacity
}
