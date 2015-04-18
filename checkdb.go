// Copyright (C) 2015 Adriano Soares <adrianosoaresjn@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Card struct {
	Id   int
	Name string
}

var setup = `
CREATE TABLE [datas] (
  [id] integer PRIMARY KEY, 
  [ot] integer, 
  [alias] integet, 
  [setcode] INT64, 
  [type] integer, 
  [atk] integer, 
  [def] integer, 
  [level] integer, 
  [race] integer, 
  [attribute] integer, 
  [category] integer);
CREATE TABLE [texts] (
	[id]	integer,
	[name]	varchar(128),
	[desc]	varchar(1024),
	[str1]	varchar(256),
	[str2]	varchar(256),
	[str3]	varchar(256),
	[str4]	varchar(256),
	[str5]	varchar(256),
	[str6]	varchar(256),
	[str7]	varchar(256),
	[str8]	varchar(256),
	[str9]	varchar(256),
	[str10]	varchar(256),
	[str11]	varchar(256),
	[str12]	varchar(256),
	[str13]	varchar(256),
	[str14]	varchar(256),
	[str15]	varchar(256),
	[str16]	varchar(256),
	PRIMARY KEY(id)
);
`

func main() {
	dbName := flag.String("db", "cards.cdb", "database name")
	flag.Parse()
	dir := flag.Arg(0)
	if dir == "" {
		fmt.Println("usage:\n\tcheckdb -db cards.cdb folder")
		os.Exit(2)
	}

	db := sqlx.MustOpen("sqlite3", *dbName)
	var cards []Card
	err := db.Select(&cards, "select id, name from texts")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(cards), "cards in", *dbName)

	tempDb := sqlx.MustOpen("sqlite3", ":memory:")
	tempDb.MustExec(setup)

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		panic(err)
	}

	fmt.Println(len(files), "sql files")
	for _, f := range files {
		file, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err)
		}
		_, err = tempDb.Exec(string(file))
		if err != nil {
			fmt.Println("tempDb.Exec error:", err)
		}
	}

	var tempDbCards []Card
	tempDb.Select(&tempDbCards, "select id, name from texts")

	cardMap := make(map[int]string)
	for _, c := range tempDbCards {
		cardMap[c.Id] = c.Name
	}

	var missingCards []string
	for _, c := range cards {
		_, found := cardMap[c.Id]
		if !found {
			missingCards = append(missingCards, fmt.Sprint(c.Id, " ", c.Name))
		}
	}
	size := len(missingCards)
	fmt.Println(size, "missing cards")
	if size == 0 {
		os.Exit(1)
	}
	if size < 20 {
		fmt.Println(strings.Join(missingCards, "\r\n"))
	} else {
		file, err := os.Create("errors.txt")
		if err != nil {
			panic(err)
		}
		_, err = file.WriteString(strings.Join(missingCards, "\r\n"))
		if err != nil {
			panic(err)
		}
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}
}
