wrpc as in Web RPC. A package that enables running any function implementing `type RemoteCall func(in io.Reader, out io.WriteCloser)
` remotely on a mesh of web workers.

Communication between the workers is done through pipes.
This allows combinations of any concurrently active pipelines. Since worker's can reschedule their work to another worker, automatic load balancing should be possible in theory.

All this is achieved by having the manager (browser main thread) and workers run the same binary and sending common function pointers defined throughout the codebase as calls are scheduled. The function must be statically declared and cannot access caller scope (where the call function is defined).

`RemoteCall` has 2 parameters: `in` and `out` as in input and output.
If main thread would want to get a result from a call to a worker, it would have to create a pair of piped ports using `wrpc.Pipe()`. One end is given to the worker into which it writes the result and from the other end we can read it back.

It is possible for the worker to create a new pipe and call subworkers and so on in any combination by just connecting the pipes together using `io.Copy` for example.

The package provides easy interfaces for launching remote calls on workers:
`Go(in io.Reader, out io.WriteCloser, f RemoteCall)` and for chaining workers by connecting each output to next input and passing `out` directly to the final worker: `GoChain(in io.Reader, out io.WriteCloser, calls ...RemoteCall)`

By having such an interface combined with the mesh network, it allows to implement any protocols on top of it. Even to go as far as to run a gRPC server as a worker call and having it schedule a call [containing the gRPC client calling the server back] to another worker. I tried a pure `net.Conn` approach at first and ran a gRPC setup on top of that. Although it worked well with `gogoproto` custom marshaling (In a raw audio application there was a noticable difference over reflection based marshaling). I instead implemented the MessagePort API directly as blocking `io.ReadWriteCloser` pipes keeping the higher abstractions open.

Demos available at: https://github.com/mgnsk/go-wasm-demos
