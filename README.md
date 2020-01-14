## Go WASM demos

### System dependencies:
* `direnv`
* `docker`
* `docker-compose`

### WASM apps
* `audiotrack` - Multi-worker audio FX chain using wrpc pipes.
* `benchmark` - Byte throughput benchmark for WebWorker chains.
* `cube` - 3D rotating cube using webgl. Use WASD and arrow keys to navigate.
* `streaming` - Unix-y text processing on wrpc pipes.
* `triangle` - 2D triangle using webgl.

Live demos: https://mgnsk.github.io/go-wasm-demos/public/
Most of the stuff except WebGL runs on dev console.

## Setup

To set up the environment, set up direnv on your system (follow the official instructions for your shell) and run:
* `$ direnv allow .` to allow loading the sandbox host environment.
* `$ sh setup.sh` to install task manager.
* `$ tusk sandbox.new` to set up and enter the sandbox shell.

Run `$ build-all` to build all wasm apps.

Finally, run `$ tusk go.serve` to serve all list of the apps at `http://localhost:8080`.

In the sandbox, to list the tasks, run: `$ tusk`. (Inside the sandbox, tusk is reconfigured to use sandbox.tusk.yml file by default).

Some commands are still scripts:
* `$ build-all` builds all wasm apps.
* `$ render-templates` renders a each app into `public` directory.
* `$ go-generate` generates all go code. (Not including protos).

### optional Visual Studio Code setup
Open the `local.code-workspace` workspace. This enables running go tools in `js,wasm` environment.
Using attached containers is also possible. After attaching, open use the `sandbox/sandbox.code-workspace`.

* Remote - Containers
https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers


## TODO
GPU accelerated WebGL tests running on headless docker would probably need special hardware.
Other tests can be run either through nodejs or wasmbrowsertest. More information: https://github.com/golang/go/wiki/WebAssembly

Headless chrome can be run but it will start complaining about
missing GL libraries if you use anything that kind.
