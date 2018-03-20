//
// Utility package for finding running containers and their details.
//

package docker

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

//
// A structure to describe a running container.
//
type DockerGuest struct {
	// Name of the image such as "example/wordpress"
	Name string

	//
	// Friendly name of that image, such as "example_wordpress"
	//
	// This is required because haproxy doesn't like all characters
	// being used in ACL or backend-names.
	//
	FriendlyName string

	//
	// The ID of the running container.
	//
	ID string

	//
	// The IP address the docker guest is listening upon
	//
	// NOTE: This is a string because we are merely parsing it
	// and pasting it into the HAProxy configuration file.  Of
	// course it is more natural to store it as an IP address..
	//
	IP string
}


//
// A function which is invoked as a callback when containers
// are started/stopped
//
type DockerCallback func()


//
// Does the given file exist?
//
func Exists(name string) bool {
    _, err := os.Stat(name)
    return !os.IsNotExist(err)
}


//
// Check that docker is installed where we expect
//
func CheckDocker() {
	if ! Exists("/usr/bin/docker") {
		fmt.Printf("/usr/bin/docker was not found!" )
		os.Exit(1)
	}
}


//
// Return details of all running containers.
//
func AllRunningContainers() ([]DockerGuest, error) {

	CheckDocker()

	//
	// These are the guests that are running
	//
	var guests []DockerGuest

	//
	// For each image which is running we'll find the name
	// of the image, and the running ID.
	//
	out, err := exec.Command("/usr/bin/docker",
		"ps", "--format", "{{.Image}},{{.ID}}").Output()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//
	// Make a Regex to say we only want alphanumeric characters
	//
	// This is used to convert the name of the image to a friendly
	// name for HAProxy.
	//
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//
	// For each running container we'll add a structure with its
	// data to our array `guests`.
	//
	entries := strings.Split(string(out), "\n")
	for _, container := range entries {
		if len(container) > 1 {

			//
			// The format is "name,id".  So split on the comma
			//
			data := strings.Split(container, ",")

			//
			// The guest
			//
			var tmp DockerGuest
			tmp.Name = data[0]
			tmp.ID = data[1]
			tmp.IP = IPFor(data[1])
			tmp.FriendlyName = reg.ReplaceAllString(tmp.Name, "_")

			//
			// Add to our list.
			//
			guests = append(guests, tmp)
		}
	}

	return guests, nil
}

//
// Find the IP address assigned to the container with the specified ID.
//
func IPFor(guest string) string {

	CheckDocker()

	out, err := exec.Command("/usr/bin/docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", guest).Output()
	if err != nil {
		panic(err)
	}
	return (strings.TrimSpace(string(out)))
}


//
// Watch for new containers being started, or existing ones being
// removed.
//
func Watch(fn DockerCallback) {
	cmd := exec.Command("/usr/bin/docker",
		"events",
		"--filter",
		"event=start",
		"--filter",
		"event=stop",
		"--filter",
		"type=container")
	out, _ := cmd.StdoutPipe()
	cmd.Start()

	rd := bufio.NewReader(out)
	for {
		_, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal("Read Error:", err)
			return
		}
		fmt.Printf("Docker event received\n")

		fn()
	}

	//
	// Not reached
	//
	// cmd.Wait()

}
