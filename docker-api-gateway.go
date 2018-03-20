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
	"flag"
	"fmt"
	"github.com/skx/docker-api-gateway/docker"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

//
// Our command-line flags
//
type CommandLineOptions struct {
	template_file string
	haproxy_file  string
	reload_cmd    string
}

//
// The flags all the routines use.
//
var FLAGS CommandLineOptions

//
// Output a configuration file for haproxy containing all the current
// guests which are present.
//
func OutputHAProxyConfig() {

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
	t, err := template.ParseFiles(FLAGS.template_file)
	if err != nil {
		log.Fatal("Failed to load template-file", err)
		return
	}

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
	err = ioutil.WriteFile(FLAGS.haproxy_file, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	//
	// Reload HAProxy
	//
	reload := strings.Split(FLAGS.reload_cmd, " ")
	_, err = exec.Command(reload[0], reload[1:]...).Output()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\tReloaded haproxy.service\n")

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
		_, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal("Read Error:", err)
			return
		}
		fmt.Printf("Event received - regenerating %s\n", FLAGS.haproxy_file)

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

	//
	// Parse our flags
	//
	flag.StringVar(&FLAGS.template_file, "template-file", "haproxy.tmpl", "The path to the haproxy template-file")
	flag.StringVar(&FLAGS.haproxy_file, "output-file", "/etc/haproxy/haproxy.cfg", "The path to haproxy configuration file to generate")
	flag.StringVar(&FLAGS.reload_cmd, "reload-command", "/bin/systemctl reload haproxy.service", "The command to execute to reload haproxy")

	flag.Parse()

	if _, err := os.Stat(FLAGS.template_file); os.IsNotExist(err) {
		fmt.Printf("The default 'haproxy.tmpl' file was not found in this directory\n")
		fmt.Printf("Please specify one; you can use '-help' to see all flags\n")
		os.Exit(1)
	}
	WatchDocker()
}
