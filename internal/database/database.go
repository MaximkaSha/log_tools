package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/MaximkaSha/log_tools/internal/utils"
	_ "github.com/lib/pq"
)

type Database struct {
	ConString string
	DB        *sql.DB
}

func NewDatabase(con string) Database {
	return Database{
		ConString: con,
	}
}

func (d *Database) InitDatabase() {
	psqlconn := d.ConString
	var err error
	d.DB, err = sql.Open("postgres", psqlconn)
	CheckError(err)
	err = d.DB.Ping()
	CheckError(err)
	log.Println("DB Connected!")
	err = d.CreateDBIfNotExist()
	CheckError(err)
	err = d.CreateTableIfNotExist()
	CheckError(err)

}

func CheckError(err error) {
	if err != nil {
		log.Printf("Database error: %s", err)
	}
}

func (d Database) CreateDBIfNotExist() error {
	var query = `SELECT 'CREATE DATABASE logs'
	WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'logs')`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err := d.DB.ExecContext(ctx, query)
	return err
}

func (d Database) CreateTableIfNotExist() error {
	var query = `CREATE TABLE IF NOT EXISTS public.log_data_2
(
    id character varying(100) COLLATE pg_catalog."default" NOT NULL,
    mtype character varying(100) COLLATE pg_catalog."default" NOT NULL,
    delta bigint,
    value double precision,
    hash character varying COLLATE pg_catalog."default",
	PRIMARY KEY (id)
)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err := d.DB.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating  table", err)
		return err
	}
	return err

}

func (d Database) InsertMetric(ctx context.Context, m models.Metrics) error {
	var query = `INSERT INTO log_data_2 (id, mtype, delta, value, hash)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id)
	DO UPDATE SET
	mtype = EXCLUDED.mtype,
	delta = EXCLUDED.delta + log_data_2.delta,
	value = EXCLUDED.value,
	hash = EXCLUDED.hash`
	_, err := d.DB.ExecContext(ctx, query, m.ID, m.MType, m.Delta, m.Value, m.Hash)
	if err != nil {
		log.Printf("Error %s when appending  data", err)
		return err
	}
	return err
}

func (d Database) GetMetric(data models.Metrics) (models.Metrics, error) {
	//log.Println(data)
	err := d.DB.QueryRow("SELECT mtype,delta,value,hash FROM log_data_2 WHERE id = $1", data.ID).Scan(&data.MType, &data.Delta, &data.Value, &data.Hash)
	//log.Println(data)
	if data.Delta == nil && data.Value == nil {
		data.Delta = new(int64)
		data.Value = new(float64)
		err = errors.New("no data")
		return data, err
	}
	return data, err

}

func (d Database) GetAll(ctx context.Context) []models.Metrics {
	var query = `SELECT * from log_data_2`
	rows, err := d.DB.QueryContext(ctx, query)
	rows.Err()
	if err != nil {
		log.Printf("Error %s when getting all  data", err)
	}
	defer rows.Close()
	data := []models.Metrics{}
	for rows.Next() {
		model := models.Metrics{}
		if err := rows.Scan(&model.ID, &model.MType, &model.Delta, &model.Value, &model.Hash); err != nil {
			log.Fatal(err)
		}
		data = append(data, model)
	}
	return data
}

func (d Database) InsertData(ctx context.Context, typeVar string, name string, value string, hash string) int {
	var model models.Metrics
	model.ID = name
	model.MType = typeVar
	if typeVar == "gauge" {
		if utils.CheckIfStringIsNumber(value) {
			tmp, _ := strconv.ParseFloat(value, 64)
			model.Value = &tmp
		} else {
			//http.Error(w, "Bad value found!", http.StatusBadRequest)
			return http.StatusBadRequest
		}
	}
	if typeVar == "counter" {
		if utils.CheckIfStringIsNumber(value) {
			tmp, _ := strconv.ParseInt(value, 10, 64)
			model.Delta = &tmp
			//	log.Println(*model.Delta)
		} else {
			//http.Error(w, "Bad value found!", http.StatusBadRequest)
			return http.StatusBadRequest
		}
	}
	model.Hash = hash
	d.InsertMetric(ctx, model)
	return http.StatusOK
}

func (d Database) SaveData(file string) {
	if file == "" {
		return
	}
	ctx := context.TODO()
	//	defer cancel()
	jData, err := json.Marshal(d.GetAll(ctx))
	if err != nil {
		log.Panic(err)
	}
	_ = ioutil.WriteFile(file, jData, 0644)
}

func (d Database) Restore(file string) {
	log.Println("DB Connected, no need to restore from file")
}

func (d Database) PingDB() bool {
	if k := d.DB.Ping(); k != nil {
		log.Println("cant ping DB!")
		return false
	}
	return true
}

func (d Database) BatchInsert(ctx context.Context, dataModels []models.Metrics) error {
	if len(dataModels) == 0 {
		return errors.New("empty batch")
	}
	for _, k := range dataModels {
		if k.ID == "RandomValue" && *k.Value == d.GetCurrentCommit() {
			return errors.New("already commited")
		}
	}
	var query = `INSERT INTO log_data_2 (id, mtype, delta, value, hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id)
		DO UPDATE SET
		mtype = EXCLUDED.mtype,
		delta = EXCLUDED.delta + log_data_2.delta,
		value = EXCLUDED.value,
		hash = EXCLUDED.hash`
	// шаг 1 — объявляем транзакцию
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()
	// шаг 2 — готовим инструкцию

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer stmt.Close()

	for _, v := range dataModels {
		// шаг 3 — указываем, что каждое видео будет добавлено в транзакцию
		if _, err = stmt.ExecContext(ctx, v.ID, v.MType, v.Delta, v.Value, v.Hash); err != nil {
			return err
		}
	}
	// шаг 4 — сохраняем изменения
	return tx.Commit()

}

func (d Database) GetCurrentCommit() float64 {
	randVal := models.Metrics{
		ID: "RandomValue",
	}
	randVal, err := d.GetMetric(randVal)
	if err != nil {
		return 0
	}
	return *randVal.Value
}
