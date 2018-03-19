<<<<<<< HEAD
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
* http://localhost/hello/world
   * Will send traffic to the running container of image hello/world on port 8000.

Our assumpions are:

* You will have started the containers you expect to be using.
* Each container is a HTTP-server which is listening on port 8000.
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
=======
# docker-api-gateway
Trivial API-gateway for docker, via HAProxy
>>>>>>> 0c8a786059a3de863c0dc127545125203b1e3bee
