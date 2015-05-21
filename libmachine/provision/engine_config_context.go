package provision

import (
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
)

type EngineConfigContext struct {
	DockerPort    int
	OtherArgs 	  string
	AuthOptions   auth.AuthOptions
	EngineOptions engine.EngineOptions
}
