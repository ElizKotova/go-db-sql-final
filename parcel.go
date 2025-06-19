package main

import (
	"database/sql"
	"errors"
)

// ParcelStore представляет хранилище посылок.
type ParcelStore struct {
	db *sql.DB
}

// NewParcelStore создает экземпляр ParcelStore.
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add создает новую посылку в БД.
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (@client, @status, @address, @created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get возвращает информацию о посылке по её номеру.
func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	err := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = @number",
		sql.Named("number", number)).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, errors.New("parcel not found")
		}
		return p, err
	}
	return p, nil
}

// GetByClient возвращает список посылок клиента.
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = @client",
		sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ps []Parcel
	for rows.Next() {
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ps, nil
}

// SetStatus обновляет статус посылки.
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = @status WHERE number = @number",
		sql.Named("status", status),
		sql.Named("number", number))
	return err
}

// SetAddress обновляет адрес посылки только если статус 'registered'.
func (s ParcelStore) SetAddress(number int, address string) error {
	res, err := s.db.Exec("UPDATE parcel SET address = @address WHERE number = @number AND status = @status",
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("parcel already sent")
	}
	return nil
}

// Delete удаляет посылку только если статус 'registered'.
func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec("DELETE FROM parcel WHERE number = @number AND status = @status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("parcel already sent")
	}
	return nil
}
