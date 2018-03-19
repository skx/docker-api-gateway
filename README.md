# Docker API Gateway

This is a simple project which is designed to allow you to setup an
API-gateway based upon a single host running docker.

In short it is assumed that you have a small number of containers and
you wish to unify their access behind a common (URI) prefix.  Traffic will
be routed to each container via the prefix of the image name.

The deamon will watch containers being launched, and terminated, by the
docker deamon and will use the information to update a global `haproxy.cfg`
file.

HAProxy will bind to `*:80` and route traffic to the appropriate
backend based on the URL prefix.  For example:

* http://localhost/foo/bar
  * Will send traffic to the running container of image foo/bar on port 8000.
  * Will send traffic to the running container of image foo/bar on port 80.
* http://localhost/hello/world
  * Will send traffic to the running container of image hello/world on port 80
  * Will send traffic to the running container of image hello/world on port 8000.

Our assumpions are:

* You will have started the containers you expect to be using.
* Each container is a HTTP-server which is listening on port 8000, or port 80.
  * These seem to be the most common.
  * Adding new ports is easy, and the HAProxy healthchecks will ensure that only the live-port will be used.
* HAProxy will handle all the routing.



## Getting Started

Assuming you have no docker guests running you should launch the
api-gateway like so:

     go run ./docker-api-gateway.go

This will start the docker gateway running, and it will listen to
events produced by docker (which new containers are launched and
terminated).

At this point nothing will be running, but you can start a simple
example by fetching the image `crccheck/hello-world` and launching
it in the background:

     root@frodo:~# docker run -d crccheck/hello-world
     df6aabd4b13363c979bbc64618a7e087e3c18c318f2eea626e8f79c84002bf0d

At this point you should notice that the file `/etc/haproxy/haproxy.cfg` has
been rewritten and will now include:

     ..

	 acl crccheck_hello_world path_beg /crccheck/hello-world
	 use_backend crccheck_hello_world-backend if crccheck_hello_world
     ..

     backend crccheck_hello_world-backend
      reqrep ^(GET|HEAD|POST)\ /crccheck/hello-world(.*) \1\ /\2
      server name 172.17.0.2:8000 check inter 2000

This means that access to this container can be achieved by hitting:

     $ curl http://localhost/crccheck/hello-world
     $ curl http://localhost/crccheck/hello-world/

If you stop the container you'll see instead that all accesses will return
a `403` response, via the default-handler.

Similarly if you spin up a PHP container like so:

      root@frodo:~# docker run -d ipunktbs/phpinfo

Now you should find PHPInfo() in all its glory:

     $ curl http://localhost/ipunktbs/phpinfo/



## HAProxy

The HAProxy configuration file is built from the template you can
find in this repository `haproxy.tmpl`.


## Disclaimer

This is a quick hack.  I've tested it with several simple HTTP-based
containers which expose themselves on :8000 and it seems to work.

I've made effort to strip dangerous characters from the haproxy configuration
file, and performed prefix-removal, via regular expressions.

Otherwise feedback is welcome.

Steve
--
