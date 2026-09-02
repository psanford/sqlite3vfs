package sqlite3vfs

/*
   #include "sqlite3vfs.h"
   #include <stdlib.h>
*/
import "C"

import (
	"io"
	"unsafe"
)

// VFSFind returns the VFS sqlite has registered under name, wrapped so that a
// Go VFS in this package can delegate to it. An empty name returns sqlite's
// current default VFS. It returns nil when no VFS is registered under that
// name.
//
// This exists for VFS implementations that cannot serve every file sqlite asks
// them for. A read-only or compressed VFS, for example, has nowhere to put the
// temp file sqlite opens when a sorter or a transient index spills, and failing
// that open fails the query. Such a VFS can pass anything that is not the main
// database down to the VFS underneath it, which is what sqlite's own appendvfs
// does:
//
//	func (vfs *MyVFS) Open(name string, flags sqlite3vfs.OpenFlag) (sqlite3vfs.File, sqlite3vfs.OpenFlag, error) {
//		if flags&sqlite3vfs.OpenMainDB == 0 {
//			return vfs.base.Open(name, flags)
//		}
//		// ... open the main database yourself
//	}
//
// The returned VFS is a handle onto a VFS sqlite already owns; there is nothing
// to close, and it must not be passed to RegisterVFS.
func VFSFind(name string) VFS {
	var cName *C.char

	if name != "" {
		cName = C.CString(name)
		defer C.free(unsafe.Pointer(cName))
	}

	vfs := C.s3vfsFind(cName)
	if vfs == nil {
		return nil
	}

	return &baseVFS{vfs: vfs}
}

// baseVFS is a VFS backed by a sqlite3_vfs registered in sqlite itself.
type baseVFS struct {
	vfs *C.sqlite3_vfs
}

var _ VFS = (*baseVFS)(nil)

// MaxPathname reports the longest path the underlying VFS accepts.
func (v *baseVFS) MaxPathname() int {
	return int(C.s3vfsBaseMaxPathname(v.vfs))
}

func (v *baseVFS) Open(name string, flags OpenFlag) (File, OpenFlag, error) {
	// A nil name (not a pointer to "") is how sqlite asks for a temp file.
	var cName *C.char

	if name != "" {
		cName = C.CString(name)
		defer C.free(unsafe.Pointer(cName))
	}

	file := C.s3vfsBaseFileAlloc(v.vfs)
	if file == nil {
		return nil, 0, NoMemError
	}

	var outFlags C.int

	rc := C.s3vfsBaseOpen(v.vfs, cName, file, C.int(flags), &outFlags)
	if int(rc) != sqliteOK {
		C.s3vfsBaseFileFree(file)

		return nil, 0, errFromCode(int(rc))
	}

	return &baseFile{file: file}, OpenFlag(outFlags), nil
}

func (v *baseVFS) Delete(name string, dirSync bool) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var syncDir C.int
	if dirSync {
		syncDir = 1
	}

	return errFromCode(int(C.s3vfsBaseDelete(v.vfs, cName, syncDir)))
}

func (v *baseVFS) Access(name string, flags AccessFlag) (bool, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var out C.int

	rc := C.s3vfsBaseAccess(v.vfs, cName, C.int(flags), &out)
	if int(rc) != sqliteOK {
		return false, errFromCode(int(rc))
	}

	return out != 0, nil
}

// FullPathname returns the name unchanged on error, since this interface has
// no way to report one.
func (v *baseVFS) FullPathname(name string) string {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	size := int(C.s3vfsBaseMaxPathname(v.vfs)) + 1
	out := (*C.char)(C.malloc(C.size_t(size)))
	defer C.free(unsafe.Pointer(out))

	rc := C.s3vfsBaseFullPathname(v.vfs, cName, C.int(size), out)
	if int(rc) != sqliteOK {
		return name
	}

	return C.GoString(out)
}

// baseFile is a File backed by a sqlite3_file belonging to another VFS.
type baseFile struct {
	file *C.sqlite3_file
}

var _ File = (*baseFile)(nil)

func (f *baseFile) Close() error {
	if f.file == nil {
		return nil
	}

	rc := C.s3vfsBaseClose(f.file)
	C.s3vfsBaseFileFree(f.file)
	f.file = nil

	return errFromCode(int(rc))
}

// ReadAt reports a short read as (len(p), io.EOF): xRead zero-fills the tail
// and keeps the bytes it read in p, so a smaller n would get them overwritten
// with zeros on the way back to sqlite.
func (f *baseFile) ReadAt(p []byte, off int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	rc := int(C.s3vfsBaseRead(f.file, unsafe.Pointer(&p[0]), C.int(len(p)), C.sqlite3_int64(off)))

	switch rc {
	case sqliteOK:
		return len(p), nil
	case IOErrorShortRead.code:
		return len(p), io.EOF
	default:
		return 0, errFromCode(rc)
	}
}

func (f *baseFile) WriteAt(p []byte, off int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	rc := int(C.s3vfsBaseWrite(f.file, unsafe.Pointer(&p[0]), C.int(len(p)), C.sqlite3_int64(off)))
	if rc != sqliteOK {
		return 0, errFromCode(rc)
	}

	return len(p), nil
}

func (f *baseFile) Truncate(size int64) error {
	return errFromCode(int(C.s3vfsBaseTruncate(f.file, C.sqlite3_int64(size))))
}

func (f *baseFile) Sync(flag SyncType) error {
	return errFromCode(int(C.s3vfsBaseSync(f.file, C.int(flag))))
}

func (f *baseFile) FileSize() (int64, error) {
	var size C.sqlite3_int64

	rc := C.s3vfsBaseFileSize(f.file, &size)
	if int(rc) != sqliteOK {
		return 0, errFromCode(int(rc))
	}

	return int64(size), nil
}

func (f *baseFile) Lock(elock LockType) error {
	return errFromCode(int(C.s3vfsBaseLock(f.file, C.int(elock))))
}

func (f *baseFile) Unlock(elock LockType) error {
	return errFromCode(int(C.s3vfsBaseUnlock(f.file, C.int(elock))))
}

func (f *baseFile) CheckReservedLock() (bool, error) {
	var out C.int

	rc := C.s3vfsBaseCheckReservedLock(f.file, &out)
	if int(rc) != sqliteOK {
		return false, errFromCode(int(rc))
	}

	return out != 0, nil
}

func (f *baseFile) SectorSize() int64 {
	return int64(C.s3vfsBaseSectorSize(f.file))
}

func (f *baseFile) DeviceCharacteristics() DeviceCharacteristic {
	return DeviceCharacteristic(C.s3vfsBaseDeviceCharacteristics(f.file))
}
