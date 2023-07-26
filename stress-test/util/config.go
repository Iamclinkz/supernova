package util

import "strconv"

var (
	SimpleWebServerPort = 9000
	SchedulerPort       = SchedulerServePortStart + 1
	SchedulerAddress    = "http://localhost:" + strconv.Itoa(SchedulerPort)
)
