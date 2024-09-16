package database

import (
	"context"
	"github.com/jackc/pgx/v4"
	"log"
	"tender-service/config"

)

var dbConn *pgx.Conn

func ConnectPostgres() (*pgx.Conn, error) {
	connStr := config.LoadConfig().PostgresConn
    if connStr == "" {
        connStr = "postgresql://your_user:your_password@192.168.0.11:5432/your_database"
        //postgres://cnrprod1725736198-team-78028:cnrprod1725736198-team-78028@rc1b-5xmqy6bq501kls4m.mdb.yandexcloud.net:6432/cnrprod1725736198-team-78028
    }

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных: %v\n", err)
		return nil, err
	}
	log.Println("Успешное подключение к базе данных")
	dbConn = conn
	return dbConn, nil
}

func CreateTables() error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS tenders (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(100) NOT NULL,
            description TEXT NOT NULL,
            service_type VARCHAR(50) NOT NULL,
            status VARCHAR(50) NOT NULL,
            organization_id UUID NOT NULL,
            creator_username_id UUID NOT NULL,
            version INT NOT NULL DEFAULT 1,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );`,
        `CREATE TABLE IF NOT EXISTS tender_history (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            tender_id UUID NOT NULL,
            name VARCHAR(100) NOT NULL,
            description TEXT NOT NULL,
            service_type VARCHAR(50) NOT NULL,
            version INT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (tender_id) REFERENCES tenders(id) ON DELETE CASCADE
        );`,
        `CREATE TABLE IF NOT EXISTS bids (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(100) NOT NULL,
            description TEXT NOT NULL,
            status VARCHAR(50) NOT NULL,
            tender_id UUID NOT NULL,
            author_type VARCHAR(50) NOT NULL,
            author_id UUID NOT NULL,
            version INT NOT NULL DEFAULT 1,
            coordination VARCHAR(50) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (tender_id) REFERENCES tenders(id) ON DELETE CASCADE
        );`,
        `CREATE TABLE IF NOT EXISTS user_decisions (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL,
            bid_id UUID NOT NULL,
            decision VARCHAR(50) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES employee(id) ON DELETE CASCADE,
            FOREIGN KEY (bid_id) REFERENCES bids(id) ON DELETE CASCADE
        );`,
        `CREATE TABLE IF NOT EXISTS feedback (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL,
            bid_id UUID NOT NULL,
            bid_feedback TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES employee(id) ON DELETE CASCADE,
            FOREIGN KEY (bid_id) REFERENCES bids(id) ON DELETE CASCADE
        );`,
        `CREATE TABLE IF NOT EXISTS bid_history (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            bid_id UUID NOT NULL,
            name VARCHAR(100) NOT NULL,
            description TEXT NOT NULL,
            version INT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (bid_id) REFERENCES bids(id) ON DELETE CASCADE
        );`,
    }
    for _, query := range queries {
        _, err := dbConn.Exec(context.Background(), query)
        if err != nil {
            log.Printf("Ошибка создания таблицы: %v\n", err)
            return err
        }
    }

    log.Println("Все таблицы успешно созданы или уже существуют.")
    return nil
}
