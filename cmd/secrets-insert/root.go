package main

import (
	"flag"
	"fmt"
	"github.com/confinale/aws-secrets-inserter/pkg/replacer"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	f := flag.String("f", "", "file that contains secrets")
	i := flag.Bool("i", false, "replace in file")
	fail := flag.Bool("fail", false, "fail if at least one secret could not be replaced")
	c := flag.String("c", ".awssecrets", "config file with relevant files to replace use instead of f")
	p := flag.String("p", "::SECRET:([^:]+):SECRET::", "regex pattern to use, you can also use PATTERN= in the config file to change for all following globs")

	flag.Parse()

	err := replacer.SetPattern(*p)

	if err != nil {
		log.Fatalf("could not parse pattern %s: %v\n", *p, err)
	}

	if *f != "" {
		Replace(*f, *fail, *i)
	} else {
		patterns, err := ioutil.ReadFile(*c)
		if err != nil {
			log.Printf("error reading file %s: %v\n", *c, err)
			return
		}
		for _, v := range strings.Split(string(patterns), "\n") {
			newP := strings.TrimSpace(v)

			if strings.HasPrefix(newP, "PATTERN=") {
				log.Print("set new pattern to ", newP[8:])
				err = replacer.SetPattern(newP[8:])
				if err != nil {
					log.Fatalf("could not parse pattern %s: %v\n", newP[8:], err)
				}
			}

			glob, err := filepath.Glob(newP)
			if err != nil {
				log.Fatalf("error with pattern %s: %v\n", newP, err)
			}

			for _, file := range glob {
				log.Print("replacing file: ", file)
				Replace(file, *fail, true)
			}
		}
	}
}

func Replace(file string, fail bool, inline bool) {
	input, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("error reading file %s: %v\n", file, err)
	}

	res, err := replacer.ReplaceAll(string(input))
	if err != nil {
		if fail {
			log.Fatalf("error replacing secrets: %v\n", err)
		} else {
			log.Printf("error replacing secrets: %v\n", err)
		}
	}

	if inline {
		stat, err := os.Stat(file)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.WriteFile(file, []byte(res), stat.Mode())
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		fmt.Print(res)
	}
}
