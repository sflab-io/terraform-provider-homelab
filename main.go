package main

import (
	"context"
	"flag"
	"log"

	"github.com/abes140377/terraform-provider-homelab/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set during the release process to the release version of the binary.
// It can be overridden during go build with: go build -ldflags="-X 'main.version=1.0.0'"
var version string = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/abes140377/homelab",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
