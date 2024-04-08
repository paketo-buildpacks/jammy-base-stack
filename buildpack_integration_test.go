package acceptance_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/paketo-buildpacks/packit/v2/pexec"
)

func testBuildpackIntegration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		mavenBuildpack         string
		jvmBuildpack           string
		syftBuildpack          string
		executableJarBuildpack string

		builderConfigFilepath string

		pack    occam.Pack
		docker  occam.Docker
		source  string
		name    string
		builder string

		image     occam.Image
		container occam.Container
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()

		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		mavenBuildpack = "gcr.io/paketo-buildpacks/maven"
		jvmBuildpack = "gcr.io/paketo-buildpacks/sap-machine"
		syftBuildpack = "gcr.io/paketo-buildpacks/syft"
		executableJarBuildpack = "gcr.io/paketo-buildpacks/executable-jar"

		source, err = occam.Source(filepath.Join("integration", "testdata", "simple_app"))
		Expect(err).NotTo(HaveOccurred())

		builderConfigFile, err := os.CreateTemp("", "builder.toml")
		Expect(err).NotTo(HaveOccurred())
		builderConfigFilepath = builderConfigFile.Name()

		_, err = fmt.Fprintf(builderConfigFile, `
[stack]
  build-image = "%s:latest"
  id = "io.buildpacks.stacks.jammy"
  run-image = "%s:latest"
`,
			stack.BuildImageID,
			stack.RunImageID,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(archiveToDaemon(stack.BuildArchive, stack.BuildImageID)).To(Succeed())
		Expect(archiveToDaemon(stack.RunArchive, stack.RunImageID)).To(Succeed())

		builder = fmt.Sprintf("builder-%s", uuid.NewString())
		logs, err := createBuilder(builderConfigFilepath, builder)
		Expect(err).NotTo(HaveOccurred(), logs)
	})

	it.After(func() {
		Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())

		Expect(docker.Image.Remove.Execute(builder)).To(Succeed())
		Expect(os.RemoveAll(builderConfigFilepath)).To(Succeed())

		Expect(docker.Image.Remove.Execute(stack.BuildImageID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(stack.RunImageID)).To(Succeed())

		Expect(os.RemoveAll(source)).To(Succeed())
	})

	it("builds an app with a buildpack", func() {
		var err error
		var logs fmt.Stringer

		image, logs, err = pack.WithNoColor().Build.
			WithPullPolicy("if-not-present").
			WithBuildpacks(
				jvmBuildpack,
				syftBuildpack,
				mavenBuildpack,
				executableJarBuildpack,
			).
			WithEnv(map[string]string{
				"BP_LOG_LEVEL": "DEBUG",
			}).
			WithBuilder(builder).
			Execute(name, source)
		Expect(err).ToNot(HaveOccurred(), logs.String)

		container, err = docker.Container.Run.
			WithEnv(map[string]string{"PORT": "8080"}).
			WithPublish("8080").
			WithPublishAll().
			Execute(image.ID)
		Expect(err).NotTo(HaveOccurred())

		Eventually(container).Should(BeAvailable())
		Eventually(container).Should(Serve(ContainSubstring("Hello World! Java version")).OnPort(8080))
	})
}

func archiveToDaemon(path, id string) error {
	skopeo := pexec.NewExecutable("skopeo")

	return skopeo.Execute(pexec.Execution{
		Args: []string{
			"copy",
			fmt.Sprintf("oci-archive://%s", path),
			fmt.Sprintf("docker-daemon:%s:latest", id),
		},
	})
}

func createBuilder(config string, name string) (string, error) {
	buf := bytes.NewBuffer(nil)

	pack := pexec.NewExecutable("pack")
	err := pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"create",
			name,
			fmt.Sprintf("--config=%s", config),
		},
	})
	return buf.String(), err
}
