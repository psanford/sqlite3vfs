package sqlite3vfs

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// recordingVFS remembers every file it opens.
type recordingVFS struct {
	*TmpVFS
	opened []*TmpFile
}

func (vfs *recordingVFS) Open(name string, flags OpenFlag) (File, OpenFlag, error) {
	f, retFlags, err := vfs.TmpVFS.Open(name, flags)
	if err != nil {
		return nil, 0, err
	}

	tf := f.(*TmpFile)
	vfs.opened = append(vfs.opened, tf)
	return tf, retFlags, nil
}

// TestFailedOpenDoesNotCloseAnotherFile opens one database successfully, then
// fails to open a second, and checks the first database's file is still open.
//
// sqlite calls xClose on a file whose xOpen failed whenever pMethods is set.
// s3vfsOpen set pMethods unconditionally, but a failed goVFSOpen never stamped
// an id into the (zeroed) s3vfsFile — so that xClose arrived with id 0 and
// closed fileMap[0]: the first file this process ever opened, out from under
// the connection that owns it. That connection then failed every read that
// missed its page cache, with errors ("SQL logic error", fts5 corruption
// reports) that pointed anywhere but here.
func TestFailedOpenDoesNotCloseAnotherFile(t *testing.T) {
	vfs := &recordingVFS{TmpVFS: newTempVFS()}

	vfsName := "recordingvfs"
	if err := RegisterVFS(vfsName, vfs); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("first.db?vfs=%s", vfsName))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}

	if len(vfs.opened) == 0 {
		t.Fatal("no file was opened for first.db")
	}
	first := vfs.opened[0]

	// mode=ro of a missing file: TmpVFS refuses the open.
	missing, err := sql.Open("sqlite3", fmt.Sprintf("file:does-not-exist.db?mode=ro&vfs=%s", vfsName))
	if err != nil {
		t.Fatal(err)
	}
	if err := missing.Ping(); err == nil {
		t.Fatal("opening a missing database in mode=ro should fail")
	}
	missing.Close()

	if _, err := first.f.Stat(); err != nil {
		t.Fatalf("first.db's file was closed by the failed open of an unrelated file: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM t`).Scan(&count); err != nil {
		t.Fatalf("first.db is broken after the failed open of an unrelated file: %v", err)
	}
}
