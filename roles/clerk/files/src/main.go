package main

import (
	"clerk/fetcher"
	"clerk/sender"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
)

type pkgs struct {
	Ami      string
	Kernel   string
	Os       []fetcher.Details
	Packages []fetcher.Details
}

func main() {
	const arch = runtime.GOARCH
	args := os.Args[1:]

	if runtime.GOOS == "windows" {
		panic("Windows is not yet supported.")
	}

	if len(args) == 0 {
		panic("Please include an AMI.")
	}

	if len(args) < 2 {
		panic("Please include a package manager.")
	}

	ami := args[0]
	if !(strings.HasPrefix(ami, "ami-")) {
		panic("Please include an AMI.")
	}

	os := fetcher.OsDetails(ami)
	allPackages := fetcher.GetPackages(args)
	data := pkgs{
		Os:       os,
		Packages: allPackages}

	content, err := json.Marshal(data)

	if err != nil {
		panic(err)
	}
	sender.GraphQL(content)
	// err = ioutil.WriteFile(ami+".json", content, 0644)
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Printf("%d packages were catalogued.", len(allPackages))
}
