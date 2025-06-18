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
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	cl := int(randSource.Int63())

	// Добавление посылок.
	p1 := getTestParcel()
	p1.Client = cl
	id1, err := store.Add(p1)
	require.NoError(t, err)
	require.NotZero(t, id1)
	p1.Number = id1

	p2 := getTestParcel()
	p2.Client = cl
	id2, err := store.Add(p2)
	require.NoError(t, err)
	require.NotZero(t, id2)
	p2.Number = id2

	// Получение списка посылок.
	ps, err := store.GetByClient(cl)
	require.NoError(t, err)
	require.Len(t, ps, 2)

	// Проверка содержимого списка.
	require.Contains(t, ps, p1)
	require.Contains(t, ps, p2)
}
