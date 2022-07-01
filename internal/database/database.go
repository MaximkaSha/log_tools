package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/MaximkaSha/log_tools/internal/utils"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123456"
	dbname   = "logs"
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

func NewDefaultDatabase() Database {
	con := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return Database{
		ConString: con,
	}
}

//fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

func (d *Database) InitDatabase() {
	psqlconn := d.ConString
	//log.Println(psqlconn)
	var err error
	d.DB, err = sql.Open("postgres", psqlconn)
	//	defer d.DB.Close()
	CheckError(err)
	err = d.DB.Ping()
	CheckError(err)
	log.Println("DB Connected!")
	err = d.CreateDBIfNotExist()
	CheckError(err)
	err = d.CreateTableIfNotExist()
	CheckError(err)
	/*var a = int64(15)
	model := models.Metrics{
		ID:    "PollCounter",
		MType: "counter",
		Delta: &a,
	}
	model2 := models.Metrics{}
	err = d.AppendMetric(model)
	CheckError(err)
	model2, err = d.GetMetric(model)
	CheckError(err)
	log.Println(*model2.Delta)
	var dd = []models.Metrics{}
	dd, err = d.GetAll()
	CheckError(err)
	log.Println(dd) */

}

func CheckError(err error) {
	if err != nil {
		log.Printf("Database error: %s", err)
	}
}

func (d *Database) Ping() error {
	if e := d.DB.Ping(); e != nil {
		log.Printf("error: %s", e)
		return errors.New("cant Ping DB")
	}
	return nil
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
    delta integer,
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

func (d Database) InsertMetric(m models.Metrics) error {
	//log.Println(d.DB.Ping())
	//log.Println("----------------------------------")
	var query = `INSERT INTO log_data_2 (id, mtype, delta, value, hash)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id)
	DO UPDATE SET
	mtype = EXCLUDED.mtype,
	delta = EXCLUDED.delta,
	value = EXCLUDED.value,
	hash = EXCLUDED.hash`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err := d.DB.ExecContext(ctx, query, m.ID, m.MType, m.Delta, m.Value, m.Hash)
	if err != nil {
		log.Printf("Error %s when appending  data", err)
		return err
	}
	return err
}

func (d Database) GetMetric(data models.Metrics) (models.Metrics, error) {
	err := d.DB.QueryRow("SELECT * FROM log_data_2 WHERE id = $1", data.ID).Scan(&data.ID, &data.MType, data.Delta, data.Value, &data.Hash)
	return data, err

}

func (d Database) GetAll() []models.Metrics {
	var query = `SELECT * from log_data_2`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when getting all  data", err)
	}
	defer rows.Close()
	//println(rows)
	data := []models.Metrics{}
	for rows.Next() {
		model := models.Metrics{}
		if err := rows.Scan(&model.ID, &model.MType, &model.Delta, &model.Value, &model.Hash); err != nil {
			log.Fatal(err)
		}
		data = append(data, model)
		//log.Printf("this is something: %v\n", *model.Delta)
	}
	return data
}

func (d Database) InsertData(typeVar string, name string, value string, hash string) int {
	var model models.Metrics
	model.ID = name
	model.MType = typeVar
	//	log.Println(value)
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
	d.InsertMetric(model)
	return http.StatusOK
}

func (d Database) SaveData(file string) {
	if file == "" {
		return
	}
	jData, err := json.Marshal(d.GetAll())
	if err != nil {
		log.Panic(err)
	}
	_ = ioutil.WriteFile(file, jData, 0644)
}

func (d Database) Restore(file string) {
	if _, err := os.Stat(file); err != nil {
		log.Println("Restore file not found")
		return
	}
	var data []models.Metrics
	var jData, err = ioutil.ReadFile(file)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal(jData, &data)
	if err != nil {
		log.Println("Data file corrupted")
	} else {
		for k := range data {
			d.InsertMetric(data[k])
		}
		log.Print("Data restored from file")
	}
}
