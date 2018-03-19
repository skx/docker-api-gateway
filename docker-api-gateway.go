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
	"github.com/skx/docker-api-gateway/docker"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"text/template"
	"time"
)

//
// Output a configuration file for haproxy containing all the current
// guests which are present.
//
func OutputHAProxyConfig(template_file string, output_file string) {

	//
	// Find all running containers
	//
	guests, err := docker.AllRunningContainers()

	if err != nil {
		log.Fatal("Failed to find running guests:", err)
		return
	}

	//
	// Data for our template output.  There might be more fields
	// in the future.
	//
	type Pagedata struct {
		Guests []docker.DockerGuest
		Time   string
	}

	//
	// Create a data-structure to pass to the template.
	//
	var x Pagedata
	x.Guests = guests
	x.Time = time.Now().Format(time.RFC3339)

	//
	// Load our template-file
	//
	t := template.Must(template.New(template_file).ParseFiles(template_file))
	buf := &bytes.Buffer{}

	//
	// Generate the rendered template.
	//
	err = t.Execute(buf, x)
	if err != nil {
		panic(err)
	}

	//
	// Write the generated output to disc
	//
	err = ioutil.WriteFile(output_file, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	//
	// Reload HAProxy
	//
	_, err = exec.Command("/bin/systemctl", "reload", "haproxy.service").Output()
	if err != nil {
		panic(err)
	}

}

//
// Watch for new containers being started, or existing ones being
// removed.
//
func WatchDocker(template_file string, output_file string) {
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
		OutputHAProxyConfig(template_file, output_file)
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
	WatchDocker("haproxy.tmpl", "/etc/haproxy/haproxy.cfg")
}
