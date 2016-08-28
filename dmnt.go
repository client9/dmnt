package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Mount represents a single mount point
type Mount struct {
	Name        string
	Source      string
	Destination string
	Driver      string
	Mode        string
	RW          bool
	Propagation string
}

// Container is a truncated representation of a Container
// (just need mount points, not everything else)
type Container struct {
	Mounts []Mount
}

// ContainerID attempts to determine the container id of the running
func ContainerID() string {
	// docker sets the hostname to 12 character hex string
	cid := os.Getenv("HOSTNAME")
	if len(cid) != 12 {
		return ""
	}
	return cid
}

func standardVolume(path string) string {
	return fmt.Sprintf("-v %s:%s", path, path)
}

func dataVolume(name, path string) string {
	return fmt.Sprintf("-v %s:%s", name, path)
}

func remappedFile(original, path string) string {
	return fmt.Sprintf("-v %s:%s", original, path)
}

// ComputeMount determines the correct docker flag to use for a given path
func ComputeMount(src string) string {
	path, err := filepath.Abs(src)
	if err != nil {
		log.Fatalf("Unable to make path from %q: %s", src, err)
	}
	cid := ContainerID()
	if cid == "" {
		return standardVolume(path)
	}
	cmd := exec.Command("docker", "inspect", cid)
	raw, err := cmd.Output()
	if err != nil {
		// error is ok here, docker might not be running
		// or docker might not be able to connect to socket
		return standardVolume(path)
	}
	containers := []Container{}
	err = json.Unmarshal(raw, &containers)
	if err != nil {
		log.Fatalf("UNable to marhsall: %s", err)
	}
	if len(containers) != 1 {
		log.Fatalf("Expected only 1 container: got %d", len(containers))
	}
	mounts := containers[0].Mounts
	for _, mnt := range mounts {
		if strings.HasPrefix(path, mnt.Destination) {
			// TODO: do not know what happens with non-local drivers
			// need to add condition?  mnt.Driver == "local"
			// TODO: check that it's hex
			if len(mnt.Name) == 64 {
				return dataVolume(mnt.Name, mnt.Destination)
			}

			// -v /host/file:/new/file
			// -v /new/file:/other/dest
			// ==> -v /host/file:/other/dest
			if len(mnt.Name) == 0 {
				return remappedFile(mnt.Source, path)
			}
		}
	}
	return standardVolume(path)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("Expected at least one arg")
	}
	outflags := make([]string, 0, len(args))
	for _, path := range args {
		outflags = append(outflags, ComputeMount(path))
	}
	fmt.Println(strings.Join(outflags, " "))
}
