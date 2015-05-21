package provision

import (
	"fmt"

	"bytes"
	"text/template"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/aliyun"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
)

func init() {
	Register("CentOS6", &RegisteredProvisioner{
		New: NewCentOS6Provisioner,
	})
}

func NewCentOS6Provisioner(d drivers.Driver) Provisioner {
	return &CentOS6Provisioner{
		GenericProvisioner{
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/sysconfig/docker",
			OsReleaseId:       "centos6",
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

type CentOS6Provisioner struct {
	GenericProvisioner
}

func (provisioner *CentOS6Provisioner) Service(name string, action pkgaction.ServiceAction) error {
	command := fmt.Sprintf("sudo service %s %s", name, action.String())

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *CentOS6Provisioner) Package(name string, action pkgaction.PackageAction) error {
	log.Debugf("%s - %s begin...", name, action)
	var (
		packageAction  string
		updateMetadata = true
	)

	switch action {
	case pkgaction.Install:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
		updateMetadata = false
	case pkgaction.Upgrade:
		packageAction = "upgrade"
	}

	// TODO: This should probably have a const
	switch name {
	case "docker":
		name = "docker-io"
	}

	if updateMetadata {
		log.Info("Updating Metadata of yum...")
		// issue yum update for metadata
		if _, err := provisioner.SSHCommand("sudo yum -y update"); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("sudo yum -y %s %s", packageAction, name)

	log.Infof("Run yum %s for %s...", packageAction, name)
	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *CentOS6Provisioner) dockerDaemonResponding() bool {
	if _, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *CentOS6Provisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions, engineOptions engine.EngineOptions) error {
	log.Debug("centos6 begin...")
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	provisioner.EngineOptions.InsecureRegistry = []string{provisioner.Driver.(*aliyun.Driver).InsecureRegistry}

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	if err := provisioner.installDocker("1.5.0"); err != nil {
		return err
	}

	if err := utils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	if err := makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	if err := configureSwarm(provisioner, swarmOptions); err != nil {
		return err
	}

	return nil
}

func (provisioner *CentOS6Provisioner) installDocker(version string) error {
	dockerExec := "/usr/bin/docker"
	log.Info("Installing Docker from yum...")
	if err := provisioner.Package("docker", pkgaction.Install); err != nil {
		return err
	}

	d := provisioner.Driver

	if version != "1.5.0" {
		log.Infof("Downloading Docker of Version %s...", version)
		cmd := fmt.Sprintf("curl -s -o %s-%s https://get.docker.com/builds/Linux/x86_64/docker-%s", dockerExec, version, version)
		if _, err := drivers.RunSSHCommandFromDriver(d, cmd); err != nil {
			return err
		}

		cmd = fmt.Sprintf("mv %s %s.bak; mv /usr/bin/docker-%s %s; chmod +x %s", dockerExec, dockerExec, version, dockerExec, dockerExec)
		if _, err := drivers.RunSSHCommandFromDriver(d, cmd); err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf("sed -r -i 's/^(other_args=.*)$/\\1\"--bridge=none\"/' %s", provisioner.DaemonOptionsFile)
	if _, err := drivers.RunSSHCommandFromDriver(d, cmd); err != nil {
		return err
	}

	log.Info("Try to start Docker daemon...")
	if err := provisioner.Service("docker", pkgaction.Start); err != nil {
		return err
	}

	return nil
}

func (provisioner *CentOS6Provisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	log.Debug("centos6 docker options...")
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `
other_args='
-H tcp://0.0.0.0:{{.DockerPort}}
-H unix:///var/run/docker.sock
{{.OtherArgs}}
--tlsverify
--tlscacert {{.AuthOptions.CaCertRemotePath}}
--tlscert {{.AuthOptions.ServerCertRemotePath}}
--tlskey {{.AuthOptions.ServerKeyRemotePath}}
{{ range .EngineOptions.Labels }}--label {{.}}
{{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}}
{{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}}
{{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}}
{{ end }}
'
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		OtherArgs:     provisioner.Driver.(*aliyun.Driver).DockerArgs,
		DockerPort:    dockerPort,
		AuthOptions:   provisioner.AuthOptions,
		EngineOptions: provisioner.EngineOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: provisioner.DaemonOptionsFile,
	}, nil
}
