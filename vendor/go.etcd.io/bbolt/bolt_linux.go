package bbolt

import (
	"syscall"
	u "github.com/YesZhen/superlog_go"
)

// fdatasync flushes written data to a file descriptor.
func fdatasync(db *DB) error {
	defer u.LogEnd(u.LogBegin("volume fdatasync"))
	return syscall.Fdatasync(int(db.file.Fd()))
}
