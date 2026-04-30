package sqlite3vfs

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestURIOpener(t *testing.T) {
	vfs := &capturingVFS{}
	vfsName := "tmpfs_uri"
	err := RegisterVFS(vfsName, vfs)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:test.db?vfs=%s&poll_interval=5s&cache_size=128&_busy_timeout=10", vfsName))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err == nil {
		t.Fatal("expected open failure from capturing VFS")
	}

	if !vfs.captured {
		t.Fatal("expected OpenURI to be called")
	}
	if vfs.name != "test.db" {
		t.Fatalf("name=%q, want %q", vfs.name, "test.db")
	}
	if value, ok := vfs.params["poll_interval"]; !ok || value != "5s" {
		t.Fatalf("poll_interval=%q ok=%v, want %q true", value, ok, "5s")
	}
	if value, ok := vfs.params["cache_size"]; !ok || value != "128" {
		t.Fatalf("cache_size=%q ok=%v, want %q true", value, ok, "128")
	}
	if value, ok := vfs.params["_busy_timeout"]; !ok || value != "10" {
		t.Fatalf("_busy_timeout=%q ok=%v, want %q true", value, ok, "10")
	}
}

type capturingVFS struct {
	name     string
	params   map[string]string
	captured bool
}

func (vfs *capturingVFS) Open(name string, flags OpenFlag) (File, OpenFlag, error) {
	return nil, flags, CantOpenError
}

func (vfs *capturingVFS) OpenURI(name string, params map[string]string, flags OpenFlag) (File, OpenFlag, error) {
	if !vfs.captured {
		vfs.name = name
		vfs.params = params
		vfs.captured = true
	}
	return nil, flags, CantOpenError
}

func (vfs *capturingVFS) Delete(name string, dirSync bool) error {
	return nil
}

func (vfs *capturingVFS) Access(name string, flags AccessFlag) (bool, error) {
	return false, nil
}

func (vfs *capturingVFS) FullPathname(name string) string {
	return name
}
