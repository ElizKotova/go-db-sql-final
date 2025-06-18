package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

// Parcel представляет информацию о посылке.
type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

// ParcelService представляет сервис для работы с посылками.
type ParcelService struct {
	store ParcelStore
}

// NewParcelService создает экземпляр ParcelService.
func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

// Register регистрирует новую посылку.
func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}

	parcel.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

// PrintClientParcels выводит список посылок клиента.
func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента №%d:\n", client)
	for i, parcel := range parcels {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt, parcel.Status)
		if i < len(parcels)-1 {
			fmt.Println()
		}
	}
	return nil
}

// NextStatus обновляет статус посылки на следующий.
func (s ParcelService) NextStatus(number int) error {
	parcel, err := s.store.Get(number)
	if err != nil {
		return err
	}

	var nextStatus string
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		return nil
	}

	fmt.Printf("У посылки № %d новый статус: %s\n", number, nextStatus)

	return s.store.SetStatus(number, nextStatus)
}

// ChangeAddress обновляет адрес посылки.
func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

// Delete удаляет посылку.
func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	// Открываем соединение с базой данных.
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// Создаём сервис для работы с посылками.
	store := NewParcelStore(db)
	service := NewParcelService(store)

	// Регистрируем новую посылку.
	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Меняем адрес посылки.
	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Меняем статус посылки.
	err = service.NextStatus(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Выводим список посылок клиента.
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Пытаемся удалить отправленную посылку.
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
	}

	// Вывод посылок клиента.
	// Предыдущая посылка не должна удалиться, т.к. её статус НЕ «зарегистрирована».
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Регистрируем новую посылку.
	p, err = service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Удаляем новую посылку.
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
	}

	// Вывод посылок клиента.
	// Здесь не должно быть последней посылки, т.к. она должна была успешно удалиться.
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}
}
