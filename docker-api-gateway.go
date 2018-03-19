//
// This is a simple Docker-centric API gateway project.
//
// What this deamon does is react to containers being launched and terminated
// by Docker.  For each container it will setup a HAProxy configuration
// entry, such that traffic can be routed to it based upon the prefix of
// the URI containing the container-name.
//
// Reacting to docker events is as simple as running `docker events` and
// parsing the output - this is made considerably simpler if you apply
// filtering such that you know you're only woken up when a container
// is launched, or terminated.  i.e. Any time you see an event you need
// to refresh & reload haproxy.
//
// Steve
// --
//
//

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
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
	// NOTE: We assume all guests bind to :8000.
	//
	// NOTE: This is a string because we are merely parsing it
	// and pasting it into the HAProxy configuration file.  Of
	// course it is more natural to store it as an IP address..
	//
	IP string
}

//
// Find the IP address assigned to the container with the specified ID.
//
func IPFor(guest string) string {

	out, err := exec.Command("/usr/bin/docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", guest).Output()
	if err != nil {
		panic(err)
	}
	return (strings.TrimSpace(string(out)))
}

//
// Output a configuration file for haproxy containing all the current
// guests which are present.
//
func OutputHAProxyConfig() {

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
	}

	//
	// Make a Regex to say we only want alphanumeric characters
	//
	// This is used to convert the name of the image to a friendly
	// name for HAProxy.
	//
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		panic(err)
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

	//
	// Data for our template output.  There might be more fields
	// in the future.
	//
	type Pagedata struct {
		Guests []DockerGuest
	}
	var x Pagedata
	x.Guests = guests

	//
	// Load our template-file
	//
	t := template.Must(template.New("haproxy.tmpl").ParseFiles("haproxy.tmpl"))
	buf := &bytes.Buffer{}
	err = t.Execute(buf, x)
	if err != nil {
		panic(err)
	}

	//
	// Write the generated template out to disc
	//
	err = ioutil.WriteFile("/etc/haproxy/haproxy.cfg", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	//
	// Reload HAProxy
	//
	out, err = exec.Command("//bin/systemctl", "reload", "haproxy.service").Output()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Reloaded HAProxy :)")
}

//
// Watch for new containers being started, or existing ones being
// removed.
//
func WatchDocker() {
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
		str, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal("Read Error:", err)
			return
		}
		fmt.Println(str)
		OutputHAProxyConfig()
	}

	//
	// Not reached
	//
	// cmd.Wait()

}

//
// Entry point.
//
func main() {
	WatchDocker()
}
