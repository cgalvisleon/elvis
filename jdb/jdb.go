package jdb

import (
	"sync"
)

/**
* Ths jdb package makes it easy to create an array of database connections
* initially to posrtgresql databases.
*	Provide a connection function, validate the existence of elements such as databases, schemas, tables, colums, index, series and users and
* it is possible to create them if they do not exist.
* Also, have a execute to sql sentences to retuns json and json array,
* that valid you result return records and how many records are returned.
**/

var (
	conn *Conn
	once sync.Once
)

type Conn struct {
	Db []*Db
}

func Load() (*Conn, error) {
	once.Do(connect)

	return conn, nil
}

func Close() error {
	for _, db := range conn.Db {
		err := db.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func DB(idx int) *Db {
	return conn.Db[idx]
}

func DBClose(idx int) error {
	return conn.Db[idx].Close()
}
