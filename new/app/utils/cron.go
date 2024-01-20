package utils

import (
	"github.com/robfig/cron/v3"
)

func IsCron表达式(cronExpr string) bool {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err := parser.Parse(cronExpr)
	return err == nil
}
