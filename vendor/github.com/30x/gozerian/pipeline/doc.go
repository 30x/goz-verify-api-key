/*
Package pipeline in Gozerian allows clients to to define and create sequential processing pipes for http requests
and their associated responses. There are a couple of key concepts to understand and use this effectively:

Pipe

A Pipe represents an intermediary (proxy) during the round trip of a HTTP request from a client to a target
and any functional activity or translations that may need to be made during processing of the request or
response.

Unlike most proxies that have a single request handler pipeline (including the built-in handlers in Go), Gozerian
takes a contrarian view that the request and response pipes should be separated in order to provide the greatest
level of control and compatibility with external systems. (For example, a primary use case of Gozerian is to work
as a plugin to the ngxin server.)

In addition to the request and response pipes, a pipe maintains a context and control system. This is accessible
by type asserting the passed http.Writer to a pipeline.ControlHolder and calling its Control() function like so:

		return func(w http.ResponseWriter, r *http.Request) {
			control := w.(pipeline.ControlHolder).Control()

Once you have a pipeline.Control, your fitting now has a variety of options available to it including logging,
error handling, flow variables, etc. See the pipeline.Control documentation for more information.

Fitting

A Fitting is simply an interface that allows for custom behavior for a single step (either request or response
or both) in a Pipe. The interface is:

	RequestHandlerFunc() http.HandlerFunc
	ResponseHandlerFunc() ResponseHandlerFunc

Either of these methods may return nil. Note that the request handler is the standard Go handler
(http.RequestHandlerFunc) while a response handler is defined as a pipeline.ResponseHandlerFunc as there is no
preexisting standard function. It is defined as:

	func(w http.ResponseWriter, r *http.Request, res *http.Response)

Die

A Die is nothing more than a factory function to create a Fitting. A Die is registered with the system by name and
is called during the initialization of a Pipe. The function signature is:

	func(config interface{}) (Fitting, error)

To register a Die with the system, you need only ensure that the Die (and Fitting, of course) is compiled into
the Go executable and call RegisterDie with a unique ID and the Die function like so:

	pipeline.RegisterDie("dump", test_util.CreateDumpFitting)

Setup & Execution:

Configuration can easily done via a YAML file. Example:

	port: 8080
	target: http://httpbin.org
	pipes:                      # pipe definitions
	  main:                     # pipe id (as registered)
	    request:                # request pipeline
	    - dump:                 # name of plugin
		dumpBody: true      # plugin-specific configuration
	    response:               # response pipeline
	    - dump:                 # name of plugin
		dumpBody: true      # plugin-specific configuration
	proxies:                    # maps host & path -> pipe
	  - host: localhost         # host
	    path: /                 # path
	    pipe: main              # pipe to use

Then, your main file need only register your Dies, open the configuration file and execute the gateway:

	pipeline.RegisterDie("dump", test_util.CreateDumpFitting)

	yamlReader, err := os.Open("main.yaml")
	if err != nil {
		fmt.Print(err)
	}

	err = go_gateway.ListenAndServe(yamlReader)
	if err != nil {
		fmt.Print(err)
	}
*/
package pipeline
