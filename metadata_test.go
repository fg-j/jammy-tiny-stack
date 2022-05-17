package acceptance_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/vacation"
	"github.com/sclevine/spec"

	. "github.com/paketo-buildpacks/jam/integration/matchers"
	. "github.com/paketo-buildpacks/packit/v2/matchers"
)

func testMetadata(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect

		// skopeo pexec.Executable
		tmpDir string
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	it("builds tiny stack", func() {
		var buildReleaseDate, runReleaseDate time.Time

		by("confirming that the build image is correct", func() {
			dir := filepath.Join(tmpDir, "build-index")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(stack.BuildArchive)
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Config.Labels).To(SatisfyAll(
				HaveKeyWithValue("io.buildpacks.stack.id", "io.buildpacks.stacks.jammy.tiny"),
				HaveKeyWithValue("io.buildpacks.stack.distro.name", "ubuntu"),
				HaveKeyWithValue("io.buildpacks.stack.distro.version", "22.04"),
				HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-buildpacks/jammy-tiny-stack"),
				HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Buildpacks"),
			))

			buildReleaseDate, err = time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			Expect(err).NotTo(HaveOccurred())
			// TODO: Why do we assert that the creation time is within 10 minutes of the
			// tests being run.
			// Expect(buildReleaseDate).To(BeTemporally("~", time.Now(), 10*time.Minute))

			Expect(image).To(SatisfyAll(
				HaveFileWithContent("/etc/group", ContainSubstring("cnb:x:1000:")),
				HaveFileWithContent("/etc/passwd", ContainSubstring("cnb:x:1001:1000::/home/cnb:/bin/bash")),
				HaveDirectory("/home/cnb"),
			))

			Expect(file.Config.User).To(Equal("1001:1000"))

			Expect(file.Config.Env).To(ContainElements(
				"CNB_USER_ID=1001",
				"CNB_GROUP_ID=1000",
				"CNB_STACK_ID=io.buildpacks.stacks.jammy.tiny",
			))

			Expect(image).To(HaveFileWithContent("/etc/gitconfig", ContainLines(
				"[safe]",
				"\tdirectory = /workspace",
				"\tdirectory = /workspace/source-ws",
				"\tdirectory = /workspace/source",
			)))

			// TODO: Do we want to make assertions about the packages installed?
			Expect(image).To(HaveFileWithContent("/var/lib/dpkg/status", SatisfyAll(
				ContainSubstring("Package: build-essential"),
				ContainSubstring("Package: ca-certificates"),
				ContainSubstring("Package: curl"),
				ContainSubstring("Package: git"),
				ContainSubstring("Package: jq"),
				ContainSubstring("Package: libgmp-dev"),
				ContainSubstring("Package: libssl3"),
				ContainSubstring("Package: libyaml-0-2"),
				ContainSubstring("Package: netbase"),
				ContainSubstring("Package: openssl"),
				ContainSubstring("Package: tzdata"),
				ContainSubstring("Package: xz-utils"),
				ContainSubstring("Package: zlib1g-dev"),
			)))
		})

		by("confirming that the run image is correct", func() {
			dir := filepath.Join(tmpDir, "run-index")
			err := os.Mkdir(dir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			archive, err := os.Open(stack.RunArchive)
			Expect(err).NotTo(HaveOccurred())
			defer archive.Close()

			err = vacation.NewArchive(archive).Decompress(dir)
			Expect(err).NotTo(HaveOccurred())

			path, err := layout.FromPath(dir)
			Expect(err).NotTo(HaveOccurred())

			index, err := path.ImageIndex()
			Expect(err).NotTo(HaveOccurred())

			indexManifest, err := index.IndexManifest()
			Expect(err).NotTo(HaveOccurred())

			Expect(indexManifest.Manifests).To(HaveLen(1))
			Expect(indexManifest.Manifests[0].Platform).To(Equal(&v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}))

			image, err := index.Image(indexManifest.Manifests[0].Digest)
			Expect(err).NotTo(HaveOccurred())

			_, err = image.ConfigFile()
			// file, err := image.ConfigFile()
			Expect(err).NotTo(HaveOccurred())

			// Expect(file.Config.Labels).To(SatisfyAll(
			// 	HaveKeyWithValue("io.buildpacks.stack.id", "io.paketo.stacks.tiny"),
			// 	HaveKeyWithValue("io.buildpacks.stack.description", "distroless-like bionic + glibc + openssl + CA certs"),
			// 	HaveKeyWithValue("io.buildpacks.stack.distro.name", "ubuntu"),
			// 	HaveKeyWithValue("io.buildpacks.stack.distro.version", "18.04"),
			// 	HaveKeyWithValue("io.buildpacks.stack.homepage", "https://github.com/paketo-buildpacks/stacks"),
			// 	HaveKeyWithValue("io.buildpacks.stack.maintainer", "Paketo Buildpacks"),
			// 	HaveKeyWithValue("io.buildpacks.stack.metadata", MatchJSON("{}")),
			// 	HaveKeyWithValue("io.buildpacks.stack.mixins", ContainSubstring(`"ca-certificates"`)),
			// 	HaveKeyWithValue("io.paketo.stack.packages", ContainSubstring(`"ca-certificates"`)),
			// 	HaveKeyWithValue("io.buildpacks.base.sbom", file.RootFS.DiffIDs[len(file.RootFS.DiffIDs)-1].String()),
			// ))

			// Expect(file.Config.Labels).NotTo(HaveKeyWithValue("io.buildpacks.stack.mixins", ContainSubstring("build:")))

			// runReleaseDate, err = time.Parse(time.RFC3339, file.Config.Labels["io.buildpacks.stack.released"])
			// Expect(err).NotTo(HaveOccurred())
			// Expect(runReleaseDate).To(BeTemporally("~", time.Now(), 10*time.Minute))

			// Expect(file.Config.User).To(Equal("1000:1000"))

			// Expect(file.Config.Env).NotTo(ContainElements(
			// 	"CNB_USER_ID=1000",
			// 	"CNB_GROUP_ID=1000",
			// 	"CNB_STACK_ID=io.paketo.stacks.tiny",
			// ))

			// Expect(image).To(SatisfyAll(
			// 	HaveFileWithContent("/etc/group", ContainSubstring("cnb:x:1000:")),
			// 	HaveFileWithContent("/etc/passwd", ContainSubstring("cnb:x:1000:1000::/home/cnb:/sbin/nologin")),
			// 	HaveDirectory("/home/cnb"),
			// ))

			// diffID, err := v1.NewHash(file.Config.Labels["io.buildpacks.base.sbom"])
			// Expect(err).NotTo(HaveOccurred())

			// layer, err := image.LayerByDiffID(diffID)
			// Expect(err).NotTo(HaveOccurred())

			// Expect(layer).To(SatisfyAll(
			// 	HaveFileWithContent(`/cnb/sbom/([a-f0-9]{8}).syft.json`, ContainSubstring("https://raw.githubusercontent.com/anchore/syft/main/schema/json/schema-2.0.2.json")),
			// 	HaveFileWithContent(`/cnb/sbom/([a-f0-9]{8}).cdx.json`, ContainSubstring(`"bomFormat": "CycloneDX"`)),
			// 	HaveFileWithContent(`/cnb/sbom/([a-f0-9]{8}).cdx.json`, ContainSubstring(`"specVersion": "1.3"`)),
			// ))

			// Expect(image).To(SatisfyAll(
			// 	HaveFile("/usr/share/doc/ca-certificates/copyright"),
			// 	HaveFile("/etc/ssl/certs/ca-certificates.crt"),
			// 	HaveFile("/var/lib/dpkg/status.d/base-files"),
			// 	HaveFile("/var/lib/dpkg/status.d/ca-certificates"),
			// 	HaveFile("/var/lib/dpkg/status.d/libc6"),
			// 	HaveFile("/var/lib/dpkg/status.d/libssl1.1"),
			// 	HaveFile("/var/lib/dpkg/status.d/netbase"),
			// 	HaveFile("/var/lib/dpkg/status.d/openssl"),
			// 	HaveFile("/var/lib/dpkg/status.d/tzdata"),
			// 	HaveDirectory("/root"),
			// 	HaveDirectory("/home/nonroot"),
			// 	HaveDirectory("/tmp"),
			// 	HaveFile("/etc/services"),
			// 	HaveFile("/etc/nsswitch.conf"),
			// ))

			// Expect(image).NotTo(HaveFile("/usr/share/ca-certificates"))

			// Expect(image).To(HaveFileWithContent("/etc/os-release", SatisfyAll(
			// 	ContainSubstring(`PRETTY_NAME="Cloud Foundry Tiny"`),
			// 	ContainSubstring(`HOME_URL="https://github.com/cloudfoundry/stacks"`),
			// 	ContainSubstring(`SUPPORT_URL="https://github.com/cloudfoundry/stacks/blob/master/README.md"`),
			// 	ContainSubstring(`BUG_REPORT_URL="https://github.com/cloudfoundry/stacks/issues/new"`),
			// )))
		})
		Expect(runReleaseDate).To(Equal(buildReleaseDate))
	})
}
