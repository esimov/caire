# Compile and run compute programs on CPU

This projects contains the compiler for turning Vulkan SPIR-V compute shaders
into binaries for arm64, arm or amd64, using
[SwiftShader](https://github.com/eliasnaur/swiftshader) with a few
modifications. A runtime implemented in C and Go is available for running the
resulting binaries.

The primary use is to support a CPU-based rendering fallback for
[Gio](https://gioui.org). In particular, the `gioui.org/shader/piet` package
contains arm, arm64, amd64 binaries for
[piet-gpu](https://github.com/linebender/piet-gpu).

# Compiling and running shaders

The `init.sh` script clones the modifed SwiftShader projects and builds it for
64-bit and 32-bit. SwiftShader is not designed to cross-compile which is why a
32-bit build is needed to compile shaders for arm.

The `example/run.sh` script demonstrates compiling and running a simple compute
program.

## Issues and contributions

See the [Gio contribution guide](https://gioui.org/doc/contribute).
