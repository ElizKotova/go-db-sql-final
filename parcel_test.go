package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// RandSource источник псевдо случайных чисел.
var randSource = rand.NewSource(time.Now().UnixNano())

// getTestParcel возвращает тестовую посылку.
func getTestParcel() Parcel {
	return Parcel{
		Client:    int(randSource.Int63()),
		Status:    ParcelStatusRegistered,
		Address:   "test address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete тестирует добавление, получение и удаление посылки.
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	p := getTestParcel()

	// Добавление посылки.
	id, err := store.Add(p)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Получение посылки.
	stored, err := store.Get(id)
	require.NoError(t, err)
	p.Number = id // Устанавливаем Number для корректного сравнения
	require.Equal(t, p, stored)

	// Удаление посылки.
	err = store.Delete(id)
	require.NoError(t, err)

	// Проверка удаления.
	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress тестирует обновление адреса посылки.
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	p := getTestParcel()

	// Добавление посылки.
	id, err := store.Add(p)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Обновление адреса.
	newAddr := "new test address"
	err = store.SetAddress(id, newAddr)
	require.NoError(t, err)

	// Проверка обновления.
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddr, stored.Address)

	// Обновление статуса.
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// Попытка обновления адреса отправленной посылки.
	err = store.SetAddress(id, newAddr)
	require.Error(t, err)
}

// TestSetStatus тестирует обновление статуса посылки.
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	p := getTestParcel()

	// Добавление посылки.
	id, err := store.Add(p)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Обновление статуса.
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// Проверка обновления.
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, stored.Status)
}

// TestGetByClient тестирует получение списка посылок клиента.
func TestGetByClient(t *testing.T) {
	// Подключение к базе данных.
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	// Создание тестовых посылок.
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// Присваиваем всем посылкам одинаковый идентификатор клиента.
	client := rand.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// Добавление посылок в базу данных.
	for i := 0; i < len(parcels); i++ {
		id, err := NewParcelStore(db).Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// Получение списка посылок клиента.
	got, err := NewParcelStore(db).GetByClient(client)
	require.NoError(t, err)
	require.Len(t, got, len(parcels))

	// Проверка содержимого списка по идентификатору посылки.
	for _, p := range got {
		want, ok := parcelMap[p.Number]
		require.True(t, ok)
		require.Equal(t, want, p)
	}
}
