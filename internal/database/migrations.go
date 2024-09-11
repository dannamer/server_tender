package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
)

func CreateStandartTables(conn *pgx.Conn) error {
	// Создание типа ENUM для организации, если он не существует
	enumQuery := `
	DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'organization_type') THEN
			CREATE TYPE organization_type AS ENUM ('IE', 'LLC', 'JSC');
		END IF;
	END $$;
	`
	_, err := conn.Exec(context.Background(), enumQuery)
	if err != nil {
		return fmt.Errorf("error creating enum type: %v", err)
	}

	// Создание таблицы employee, если она не существует
	employeeTable := `
	CREATE TABLE IF NOT EXISTS employee (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		username VARCHAR(50) UNIQUE NOT NULL,
		first_name VARCHAR(50),
		last_name VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = conn.Exec(context.Background(), employeeTable)
	if err != nil {
		return fmt.Errorf("error creating employee table: %v", err)
	}

	// Создание таблицы organization, если она не существует
	organizationTable := `
	CREATE TABLE IF NOT EXISTS organization (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		type organization_type,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = conn.Exec(context.Background(), organizationTable)
	if err != nil {
		return fmt.Errorf("error creating organization table: %v", err)
	}

	// Создание таблицы organization_responsible, если она не существует
	organizationResponsibleTable := `
	CREATE TABLE IF NOT EXISTS organization_responsible (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
		user_id UUID REFERENCES employee(id) ON DELETE CASCADE
	);
	`
	_, err = conn.Exec(context.Background(), organizationResponsibleTable)
	if err != nil {
		return fmt.Errorf("error creating organization_responsible table: %v", err)
	}

	log.Println("Tables created or already exist")
	return nil
}


func CreatePrivateTables(conn *pgx.Conn) error {
	// Создание таблицы для тендеров
	tenderTable := `
	CREATE TABLE IF NOT EXISTS tender (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		service_type VARCHAR(50),
		status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
		organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
		creator_id UUID REFERENCES employee(id) ON DELETE CASCADE,
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := conn.Exec(context.Background(), tenderTable)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы тендеров: %v", err)
	}

	// Создание таблицы для предложений (bids)
	bidTable := `
	CREATE TABLE IF NOT EXISTS bid (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		description TEXT,
		status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
		tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
		organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
		creator_id UUID REFERENCES employee(id) ON DELETE CASCADE,
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = conn.Exec(context.Background(), bidTable)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы предложений: %v", err)
	}

	// Создание таблицы для истории тендеров
	tenderHistoryTable := `
	CREATE TABLE IF NOT EXISTS tender_history (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tender_id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		description TEXT NOT NULL,
		service_type VARCHAR(50) NOT NULL,
		version INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = conn.Exec(context.Background(), tenderHistoryTable)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы истории тендеров: %v", err)
	}

	log.Println("Таблицы созданы или уже существуют")
	return nil
}
