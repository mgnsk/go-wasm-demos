## Go WASM demos

### WASM apps

- `audiotrack` - Multi-worker audio FX chain using wrpc pipes.
- `benchmark` - Byte throughput benchmark for WebWorker chains.
- `cube` - 3D rotating cube using webgl. Use WASD and arrow keys to navigate.
- `httpserver` - A HTTP server running on WebWorker with client on main thread.
- `streaming` - Unix-y text processing on wrpc pipes.
- `triangle` - 2D triangle using webgl.

`direnv` is required to load the environment. Install [tusk](https://github.com/rliebz/tusk) to run tasks.
To get started, run:

- `$ tusk generate`
- `$ tusk build`
- `$ tusk serve`
