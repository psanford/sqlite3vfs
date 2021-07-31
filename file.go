package sqlite3vfs

type File interface {
	Close() error

	// ReadAt reads len(p) bytes into p starting at offset off in the underlying input source.
	// It returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
	// If n < len(p), SQLITE_IOERR_SHORT_READ will be returned to sqlite.
	ReadAt(p []byte, off int64) (n int, err error)

	// WriteAt writes len(p) bytes from p to the underlying data stream at offset off.
	// It returns the number of bytes written from p (0 <= n <= len(p)) and any error encountered that caused the write to stop early.
	// WriteAt must return a non-nil error if it returns n < len(p).
	WriteAt(p []byte, off int64) (n int, err error)

	Truncate(size int64) error

	Sync(flag SyncType) error

	FileSize() (int64, error)

	// Lock increases lock type
	Lock(elock LockType) error
	// Unlock decreases lock type
	Unlock(elock LockType) error
	// Check whether any database connection, either in this process or in some other process, is holding a RESERVED, PENDING, or EXCLUSIVE lock on the file. It returns true if such a lock exists and false otherwise.
	CheckReservedLock() (bool, error)
	// SectorSize returns the sector size of the device that underlies the file. The sector size is the minimum write that can be performed without disturbing other bytes in the file.
	SectorSize() int64
	// DeviceCharacteristics returns a bit vector describing behaviors of the underlying device.
	DeviceCharacteristics() DeviceCharacteristic
}

type SyncType int

const (
	SyncNormal   SyncType = 0x00002
	SyncFull     SyncType = 0x00003
	SyncDataOnly SyncType = 0x00010
)

// https://www.sqlite.org/c3ref/c_lock_exclusive.html
type LockType int

const (
	LockNone      LockType = 0
	LockShared    LockType = 1
	LockReserved  LockType = 2
	LockPending   LockType = 3
	LockExclusive LockType = 4
)

// https://www.sqlite.org/c3ref/c_iocap_atomic.html
type DeviceCharacteristic int

const (
	IocapAtomic              DeviceCharacteristic = 0x00000001
	IocapAtomic512           DeviceCharacteristic = 0x00000002
	IocapAtomic1K            DeviceCharacteristic = 0x00000004
	IocapAtomic2K            DeviceCharacteristic = 0x00000008
	IocapAtomic4K            DeviceCharacteristic = 0x00000010
	IocapAtomic8K            DeviceCharacteristic = 0x00000020
	IocapAtomic16K           DeviceCharacteristic = 0x00000040
	IocapAtomic32K           DeviceCharacteristic = 0x00000080
	IocapAtomic64K           DeviceCharacteristic = 0x00000100
	IocapSafeAppend          DeviceCharacteristic = 0x00000200
	IocapSequential          DeviceCharacteristic = 0x00000400
	IocapUndeletableWhenOpen DeviceCharacteristic = 0x00000800
	IocapPowersafeOverwrite  DeviceCharacteristic = 0x00001000
	IocapImmutable           DeviceCharacteristic = 0x00002000
	IocapBatchAtomic         DeviceCharacteristic = 0x00004000
)
