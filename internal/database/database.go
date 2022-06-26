package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

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
	DB        sql.DB
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
	log.Println(psqlconn)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
	d.DB = *db
	//log.Println(d.DB)

	// close database
	//defer d.DB.Close()

	// check db
	err = d.DB.Ping()
	CheckError(err)

	log.Println("DB Connected!")
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func (d *Database) Ping() error {
	if e := d.DB.Ping(); e != nil {
		log.Printf("error: %s", e)
		return errors.New("cant Ping DB")
	}
	return nil
}
