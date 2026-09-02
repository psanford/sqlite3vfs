package sqlite3vfs

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestVFSFindDefault(t *testing.T) {
	if VFSFind("") == nil {
		t.Fatal("VFSFind(\"\") returned nil; sqlite always has a default VFS")
	}

	if vfs := VFSFind("no-such-vfs"); vfs != nil {
		t.Fatalf("VFSFind of an unregistered name returned %v, want nil", vfs)
	}
}

// noTempVFS cannot make a temp file of its own and hands that open to base;
// with a nil base the open fails, as it did before VFSFind existed.
type noTempVFS struct {
	*TmpVFS
	base VFS
}

func (vfs *noTempVFS) Open(name string, flags OpenFlag) (File, OpenFlag, error) {
	if name == "" {
		if vfs.base == nil {
			return nil, 0, CantOpenError
		}

		return vfs.base.Open(name, flags)
	}

	return vfs.TmpVFS.Open(name, flags)
}

// TestDelegateTempFiles runs a query that must spill to a temp file against a
// VFS that cannot make one: it fails without a base VFS to delegate to and
// succeeds with one.
func TestDelegateTempFiles(t *testing.T) {
	for _, tc := range []struct {
		name      string
		base      VFS
		wantError bool
	}{
		{name: "no_delegate", base: nil, wantError: true},
		{name: "delegate_to_default", base: VFSFind("")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vfsName := "notemp_" + tc.name
			if err := RegisterVFS(vfsName, &noTempVFS{TmpVFS: newTempVFS(), base: tc.base}); err != nil {
				t.Fatal(err)
			}

			db, err := sql.Open("sqlite3", fmt.Sprintf("spill.db?vfs=%s", vfsName))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			// One connection so the PRAGMAs apply where the query runs;
			// temp_store=FILE plus a tiny cache forces the sort to spill.
			db.SetMaxOpenConns(1)

			for _, pragma := range []string{"PRAGMA temp_store = FILE", "PRAGMA cache_size = -16"} {
				if _, err := db.Exec(pragma); err != nil {
					t.Fatalf("%s: %s", pragma, err)
				}
			}

			if _, err := db.Exec(`CREATE TABLE spill (id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
				t.Fatal(err)
			}

			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}

			stmt, err := tx.Prepare(`INSERT INTO spill (v) VALUES (?)`)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 20000; i++ {
				if _, err := stmt.Exec(fmt.Sprintf("row-%d-padded-padded-padded-padded", i)); err != nil {
					t.Fatal(err)
				}
			}

			stmt.Close()

			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}

			rows, err := db.Query(`SELECT v, COUNT(*) FROM spill GROUP BY v ORDER BY 2, 1`)
			if err == nil {
				for rows.Next() {
				}
				err = rows.Err()
				rows.Close()
			}

			if tc.wantError {
				if err == nil {
					t.Fatal("spilling query succeeded with no VFS to open a temp file on")
				}

				t.Logf("without delegation: %s", err)

				return
			}

			if err != nil {
				t.Fatalf("spilling query failed with delegation in place: %s", err)
			}
		})
	}
}
