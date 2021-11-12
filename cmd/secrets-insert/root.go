package main

import (
	"flag"
	"fmt"
	"github.com/confinale/aws-secrets-inserter/pkg/replacer"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	f := flag.String("f", "", "file that contains secrets")
	i := flag.Bool("i", false, "replace in file")
	fail := flag.Bool("fail", false, "fail if at least one secret could not be replaced")

	flag.Parse()

	if *f == "" {
		log.Fatalln("a file has to be provided with f flag")
	}

	input, err := ioutil.ReadFile(*f)
	if err != nil {
		log.Fatalf("error reading file %s: %v\n", *f, err)
	}

	res, err := replacer.ReplaceAll(string(input))
	if err != nil && *fail {
		log.Fatalf("error replacing secrets: %v\n", err)
	}

	if *i {
		stat, err := os.Stat(*f)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.WriteFile(*f, []byte(res), stat.Mode())
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		fmt.Print(res)
	}
}
