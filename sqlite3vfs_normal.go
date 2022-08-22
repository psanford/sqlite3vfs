package sqlite3vfs

/*
   #cgo linux LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
   #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
*/
import "C"
