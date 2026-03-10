package main

import "embed"

// ApisFS contains all API YAML definition files, embedded at build time.
// Adding a new API only requires adding a YAML file under apis/ and rebuilding.
//
//go:embed apis
var ApisFS embed.FS
