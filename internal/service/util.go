package service

import (
	"github.com/lib/pq"
)

func isUniqueViolation(err error) bool {
	pqerr, ok := err.(*pq.Error)
	return ok && pqerr.Code == "23505"
}
