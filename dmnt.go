package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Mount struct {
	Name        string
	Source      string
	Destination string
	Driver      string
	Mode        string
	RW          bool
	Propagation string
}

type Container struct {
	Mounts []Mount
}

func ContainerID() string {
	return os.Getenv("HOSTNAME")
}

func StandardVolume(path string) string {
	return fmt.Sprintf("-v %s:%s", path, path)
}

func DataVolume(name, path string) string {
	return fmt.Sprintf("-v %s:%s", name, path)
}

func RemappedFile(original, path string) string {
	return fmt.Sprintf("-v %s:%s", original, path)
}

func ComputeMount(path string) string {
	cid := ContainerID()
	if cid == "" {
		return StandardVolume(path)
	}
	cmd := exec.Command("docker", "inspect", cid)
	raw, err := cmd.Output()
	if err != nil {
		return StandardVolume(path)
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
				return DataVolume(mnt.Name, mnt.Destination)
			}

			// -v /host/file:/new/file
			// -v /new/file:/other/dest
			// ==> -v /host/file:/other/dest
			if len(mnt.Name) == 0 {
				return RemappedFile(mnt.Source, path)
			}
		}
	}
	return StandardVolume(path)
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
