# otel-issue-10031

To reproduce:

1.  Install a Go toolchain.

1.  Clone this repository.

        git clone git@github.com:SeanPMiller/otel-issue-10031.git

1.  Change working directory into your local clone of this repository.

        cd otel-issue-10031

1.  Compile the binary.

        go build -v ./cmd/issue10031/ 

1.  Run the binary against the given configuration file.

        ./issue10031 --config=file:$(pwd)/issue10031.yaml
