package demo

import "os"

type EnvironmentVariable struct {
	CronExpression string
	License        string
}

func LoadConfig() EnvironmentVariable {
	return EnvironmentVariable{
		CronExpression: os.Getenv("CRON_EXPRESSION"),
		License:        os.Getenv("LICENSE"),
	}
}
