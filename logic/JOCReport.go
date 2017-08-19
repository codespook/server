package logic

import (
	"errors"
	"github.com/impactasaurus/server"
	"github.com/impactasaurus/server/auth"
	"time"
)

func GetJOCServiceReport(start, end time.Time, questionSetID string, u auth.User) (*server.JOCServiceReport, error) {
	return nil, errors.New("Not implemented")
}
