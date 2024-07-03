package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/jfrog/terraform-provider-mission-control/pkg/missioncontrol"
)

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name missioncontrol

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/jfrog/mission-control",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), missioncontrol.NewProvider(), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
