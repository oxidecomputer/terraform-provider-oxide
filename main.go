// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"flag"
	"log"

	"github.com/oxidecomputer/terraform-provider-oxide/oxide"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	if err := providerserver.Serve(
		context.Background(),
		func() provider.Provider { return oxide.New(oxide.Version) },
		providerserver.ServeOpts{
			Address: "registry.terraform.io/oxidecomputer/oxide",
			Debug:   debug,
		},
	); err != nil {
		log.Fatal(err.Error())
	}
}
