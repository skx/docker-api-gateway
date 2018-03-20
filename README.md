# Docker API Gateway

This is a simple project which is designed to allow you to setup an
API/gateway-host based upon a single machine running docker.

* You have a small number of services, or microservices, deployed as docker images.
* You wish to allow them to be reached via http://api.example.com/$app
* You don't want to worry about port-forwarding, access-control, and you don't mind launching the containers themselves by hand.

The docker-api-gateway deamon will react to containers being launched/terminated by docker, and update a global `haproxy.cfg` configuration file - which [haproxy](https://www.haproxy.org) will use.

The haproxy.cfg file, generated from [a template](haproxy.tmpl), will bind to `*:80` and route traffic to the appropriate backend based on the path requested - it is assumed the prefix will match the containers name.

* [http://localhost/foo/bar](http://localhost/foo/bar)
  * Will send traffic to the running container of image `foo/bar` on port 8000.
  * Will send traffic to the running container of image `foo/bar` on port 80.
* [http://localhost/hello/world](http://localhost/hello/world)
  * Will send traffic to the running container of image `hello/world` on port 80
  * Will send traffic to the running container of image `hello/world` on port 8000.

Our assumpions are:

* You will have started the containers you expect to be using.
* Each container will host a HTTP-server which is listening on port 8000, or port 80.
  * These seem to be the most commonly used ports, adding additional ones is not hard as they can be duplicated - The HAProxy healthchecks will ensure that only the live-port will be used.



## Getting Started

Assuming you have no docker guests running you should launch the api-gateway like so:

     go run ./docker-api-gateway.go

This will start the docker gateway running, and it will listen to events produced by docker (containers being launched or terminated).

At this point nothing will be running, but you can start a simple example by fetching the image `crccheck/hello-world` and launching it in the background:

     root@frodo:~# docker run -d crccheck/hello-world
     df6aabd4b13363c979bbc64618a7e087e3c18c318f2eea626e8f79c84002bf0d

Once the image has downloaded it will be launched, and the `docker-api-gateway` process will notice that a new container has been created.  Because a new container has been created the file `/etc/haproxy/haproxy.cfg` will be updated to include details of all the local instances.

The `/etc/haproxy/haproxy.cfg` file will now look something like this:

     ..

	 acl crccheck_hello_world path_beg /crccheck/hello-world
	 use_backend crccheck_hello_world-backend if crccheck_hello_world

     ..

     backend crccheck_hello_world-backend
      reqrep ^(GET|HEAD|POST)\ /crccheck/hello-world(.*) \1\ /\2
      server name 172.17.0.2:80   check inter 2000
      server name 172.17.0.2:8000 check inter 2000

The first section defined a match based upon the path component of the URL, and the second section sends that to the IP of the docker guest - notice we're explicitly not setting up any port-forwarding.

You can now view the container's output via:

     $ curl http://localhost/crccheck/hello-world
     $ curl http://localhost/crccheck/hello-world/

If you stop the container you'll find that the backend, and ACL, will be removed from the `haproxy.cfg` file, and that all accesses will return a `403` response, via the default-handler.

As a second-test you can spin up a different container, hosting PHP, via:

      root@frodo:~# docker run -d ipunktbs/phpinfo

Now you should find PHPInfo() in all its glory:

     $ curl http://localhost/ipunktbs/phpinfo/

You'll notice in both cases that the request made to the docker-container will have the image-name prefix stripped off it.


## Command-Line Flags

The following flags are supported:

* `-template-file`
  * The name of the template file to read to generate the output.  Default `haproxy.tmpl`.
* `-output-file`
   * The name of the haproxy configuration file to generate.  Default `/etc/haproxy/haproxy.cfg`
* `-reload-cmd`
   * The command to execute when the output file has been written.  Default `/bin/systemctl reload haproxy.service`.


## Possible Enhancements?

The most obvious way to change/improve this project is to switch from
prefix-based routing to vhost-based.

For example rather than:

* http://api.example.com/foo/bar

Allow:

* http://bar.api.example.com

Although `bar` isn't a complete identifier it is unlikely you'd have
multiple containers running with the same suffix (I assume!)  Doing this
would only involve rewriting the haproxy.cfg template - perhaps adding
a command-line flag to allow the user to choose which template to use
would make that easier.


## Disclaimer

This is a quick hack.  I've tested it with several simple HTTP-based
containers which expose themselves on `:8000` and it works well, but I don't
expect it to be universally useful as-is.

Feedback is welcome, whether good or bad :)

Steve
--
