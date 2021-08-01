# sqlite3vfs: Go sqlite3 VFS API

sqlite3vfs is a Cgo API that allows you to create custom sqlite Virtual File Systems (VFS) in Go. You can use this with the `sqlite3` https://github.com/mattn/go-sqlite3 SQL driver.


## Basic usage

To use, simply implement the [VFS](https://pkg.go.dev/github.com/psanford/sqlite3vfs?utm_source=godoc#VFS) and [File](https://pkg.go.dev/github.com/psanford/sqlite3vfs?utm_source=godoc#File) interfaces. Then register your VFS with sqlite3vfs and include the vfs name when opening the database connection:

```
	// create your VFS
	vfs := newTempVFS()

	vfsName := "tpmfs"
	err := sqlite3vfs.RegisterVFS(vfsName, vfs)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("foo.db?vfs=%s", vfsName))
	if err != nil {
		panic(err)
	}

```

A full example can be found in [sqlite3vfs_test.go](sqlite3vfs_test.go).

## Users

- [sqlite3vfshttp](https://github.com/psanford/sqlite3vfshttp): Query a sqlite database over HTTP using range requests
