package sqlite3vfs

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"
)

func TestSqlite3vfs(t *testing.T) {

	vfs := newTempVFS()

	vfsName := "tmpfs"
	err := RegisterVFS(vfsName, vfs)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("foo.db?vfs=%s", vfsName))
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS foo (
id text NOT NULL PRIMARY KEY,
title text
)`)
	if err != nil {
		t.Fatal(err)
	}

	rows := []FooRow{
		{
			ID:    "415",
			Title: "romantic-swell",
		},
		{
			ID:    "610",
			Title: "ironically-gnarl",
		},
		{
			ID:    "768",
			Title: "biophysicist-straddled",
		},
	}

	for _, row := range rows {
		_, err = db.Exec(`INSERT INTO foo (id, title) values (?, ?)`, row.ID, row.Title)
		if err != nil {
			t.Fatal(err)
		}
	}

	rowIter, err := db.Query(`SELECT id, title from foo order by id`)
	if err != nil {
		t.Fatal(err)
	}

	var gotRows []FooRow

	for rowIter.Next() {
		var row FooRow
		err = rowIter.Scan(&row.ID, &row.Title)
		if err != nil {
			t.Fatal(err)
		}
		gotRows = append(gotRows, row)
	}
	err = rowIter.Close()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(rows, gotRows) {
		t.Fatal(cmp.Diff(rows, gotRows))
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	// reopen db
	db, err = sql.Open("sqlite3", fmt.Sprintf("foo.db?vfs=%s", vfsName))
	if err != nil {
		t.Fatal(err)
	}

	rowIter, err = db.Query(`SELECT id, title from foo order by id`)
	if err != nil {
		t.Fatal(err)
	}

	gotRows = gotRows[:0]

	for rowIter.Next() {
		var row FooRow
		err = rowIter.Scan(&row.ID, &row.Title)
		if err != nil {
			t.Fatal(err)
		}
		gotRows = append(gotRows, row)
	}
	err = rowIter.Close()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(rows, gotRows) {
		t.Fatal(cmp.Diff(rows, gotRows))
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

type FooRow struct {
	ID    string
	Title string
}

func TestFileControlPragma(t *testing.T) {
	vfs := newTempVFSWithPragma()

	vfsName := "tmpfs_pragma"
	err := RegisterVFS(vfsName, vfs)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("pragma_test.db?vfs=%s", vfsName))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`PRAGMA test_pragma = 'hello_world'`)
	if err != nil {
		t.Fatal(err)
	}

	var result string
	err = db.QueryRow(`PRAGMA test_pragma`).Scan(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != "hello_world" {
		t.Fatalf("expected 'hello_world', got '%s'", result)
	}

	_, err = db.Exec(`PRAGMA test_pragma = 'updated_value'`)
	if err != nil {
		t.Fatal(err)
	}

	err = db.QueryRow(`PRAGMA test_pragma`).Scan(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != "updated_value" {
		t.Fatalf("expected 'updated_value', got '%s'", result)
	}
}
