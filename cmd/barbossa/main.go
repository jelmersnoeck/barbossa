package main

import (
	"github.com/jelmersnoeck/barbossa/internal/webhooks"

	"github.com/openshift/generic-admission-server/pkg/cmd"
)

func main() {
	cmd.RunAdmissionServer(
		&webhooks.HighAvailabilityAdmissionHook{},
	)
}
