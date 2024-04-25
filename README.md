# otel-issue-10031

To reproduce:

1.  Install a Go toolchain.

2.  Clone this repository.

        git clone git@github.com:SeanPMiller/otelcol-issue-10031.git

3.  Change working directory into your local clone of this repository.

        cd otelcol-issue-10031

4.  Compile the binary.

        go build -v ./cmd/issue10031/ 

5.  Run the binary against the given configuration file.

        ./issue10031 --config=file:$(pwd)/issue10031.yaml

6.  Observe failure.

7.  Edit the `go.mod` file, using your editor to change `v0.98.0` to `v0.97.0`.

8.  Repeat steps 4-5. You may need to run a `go get` to pull older modules.

9.  Observe success.
