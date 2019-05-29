package functions

import (
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/niranjan94/bifrost/config"
	"github.com/niranjan94/bifrost/utils"
	"github.com/niranjan94/bifrost/utils/debug"
	"github.com/niranjan94/bifrost/utils/docker"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

var containerReferences = make(map[string]*docker.Container)

const containerBase = "/cwd"

// getContainerFor returns a reference to a running docker container for the given runtime
// containers are created only when needed and are re-used if already present
func getContainerFor(runtime string) (*docker.Container, error) {
	if container, ok := containerReferences[runtime]; ok {
		return container, nil
	}
	logrus.Infof("preparing %s build environment", runtime)
	image := fmt.Sprintf("docker.io/lambci/lambda:build-%s", runtime)

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: utils.GetCwd(),
			Target: containerBase,
		},
	}

	container, err := docker.StartContainer(image, []string{"tail"}, []string{"-f", "/dev/null"}, mounts)
	if err != nil {
		return nil, err
	}
	containerReferences[runtime] = container
	return container, nil
}

// getDirectories returns a set of directories related to the given cwd working directory
// they can also be created if not present by passing makeDirectories as true
func getDirectories(cwd string, makeDirectories bool) (rootDir string, buildDir string, packageDir string) {
	var join func(...string) string

	if makeDirectories {
		join = filepath.Join
	} else {
		join = path.Join
	}

	rootDir = join(cwd, config.GetString("serverless.rootDir"))
	buildDir = join(rootDir, config.GetString("serverless.package.BuildDir"))
	packageDir = join(buildDir, "packages")

	if makeDirectories {
		os.RemoveAll(buildDir)
		os.MkdirAll(buildDir, 0751)
		os.MkdirAll(packageDir, 0751)
	}
	return rootDir, buildDir, packageDir
}

// BuildScriptInput holds are the required data to generate the build script from the buildScriptTemplate
type BuildScriptInput struct {
	RootDir          string
	BuildDir         string
	BuildPath        string
	PackageDir       string
	SourcePath       string
	PackageFile      string
	RequirementsFile string

	ShouldCleanup bool

	GlobalRequirements []string
	GlobalIncludes     []string
}

// buildScriptTemplate is the template for the build script that runs within the container
const buildScriptTemplate string = `
#!/bin/sh

rm -rf  {{.BuildPath}}
cp -rf {{.SourcePath}} {{.BuildDir}}
cd {{.BuildPath}} && pip install -r {{.RequirementsFile}} -t .
{{$BuildPath := .BuildPath}}

{{range .GlobalRequirements}}
	cd {{$BuildPath}} && pip install -r {{.}} -t .
{{end}}

{{range .GlobalIncludes}}
	cp -rf {{.}} {{$BuildPath}}
{{end}}

cd {{.BuildPath}} && zip -r9 {{.PackageFile}} .

{{if .ShouldCleanup}}
	rm -rf {{.BuildPath}}
	rm -- "$0"
{{end}}
`

// Build starts the build process for all of the serverless functions
func Build() []*DeploymentPackage {
	defer func() {
		logrus.Debug("cleaning up containers")
		for _, c := range containerReferences {
			if err := c.Remove(true); err != nil {
				logrus.Error(err)
			}
		}
	}()

	functionsMap := config.GetStringMapSub("serverless.functions", true)

	input := BuildScriptInput{
		ShouldCleanup: config.GetBool("serverless.package.cleanup"),
	}

	input.RootDir, input.BuildDir, input.PackageDir = getDirectories(containerBase, false)
	_, localBuildDir, localPackageDir := getDirectories(utils.GetCwd(), true)

	for _, name := range config.GetStringSlice("serverless.package.GlobalRequirements") {
		input.GlobalRequirements = append(input.GlobalRequirements, path.Join(input.RootDir, name))
	}

	for _, name := range config.GetStringSlice("serverless.package.GlobalIncludes") {
		input.GlobalIncludes = append(input.GlobalIncludes, path.Join(input.RootDir, name, "*"))
	}

	buildScriptTemplate := template.Must(template.New("buildScriptTemplate").Parse(buildScriptTemplate))

	namePrefix := config.GetString("serverless.prefix")
	nameSuffix := config.GetString("serverless.suffix")

	var deploymentPackages []*DeploymentPackage

	for name, function := range functionsMap {

		function.SetDefault("prefix", namePrefix)
		function.SetDefault("suffix", nameSuffix)

		functionName := function.GetString("prefix") + name + function.GetString("suffix")

		function.SetDefault("source", name)

		packageName := name + ".zip"

		input.SourcePath = path.Join(input.RootDir, function.GetString("source"))
		input.BuildPath = path.Join(input.BuildDir, path.Base(input.SourcePath))
		input.RequirementsFile = path.Join(input.SourcePath, config.GetString("serverless.package.RequirementsFile"))
		input.PackageFile = path.Join(input.PackageDir, packageName)

		var buildScript bytes.Buffer
		if err := buildScriptTemplate.Execute(&buildScript, input); err != nil {
			logrus.Error(err)
			continue
		}

		container, err := getContainerFor(function.GetString("runtime"))
		if err != nil {
			logrus.Error(err)
			continue
		}

		buildScriptFilePath := filepath.Join(localBuildDir, "build.sh")
		os.Remove(buildScriptFilePath)
		buildScriptFile, err := os.OpenFile(buildScriptFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0711)
		if err != nil {
			logrus.Error(err)
			continue
		}

		if _, err = buildScriptFile.Write(buildScript.Bytes()); err != nil {
			logrus.Error(err)
			continue
		}

		buildScriptFile.Close()

		logrus.Info("building function ", name)
		if output, err := container.RunCommand([]string{"/bin/sh", path.Join(input.BuildDir, "build.sh")}); err != nil {
			logrus.Error(err)
			debug.PrintMultilineOutput(output)
		}
		logrus.Info("built function ", name)

		deploymentPackages = append(deploymentPackages, &DeploymentPackage{
			Name:         name,
			FunctionName: functionName,
			PackageFile:  filepath.Join(localPackageDir, packageName),
			Config:       function,
		})
	}
	return deploymentPackages
}
