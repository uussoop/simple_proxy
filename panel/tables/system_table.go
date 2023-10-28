package tables

import (
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
)

type SystemTable struct {
	conn db.Connection
	c    *config.Config
}

func NewSystemTable(conn db.Connection, c *config.Config) *SystemTable {
	return &SystemTable{conn: conn, c: c}
}
func (s *SystemTable) connection() *db.SQL {
	return db.WithDriver(s.conn)
}

func (s *SystemTable) table(table string) *db.SQL {
	return s.connection().Table(table)
}
