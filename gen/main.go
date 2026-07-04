package main

import (
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const pullRef = "alpine:latest"
const saveRef = "composefork/alpine"

func main() {
	for _, arch := range []string{"amd64", "arm64"} {
		img, err := crane.Pull(pullRef, crane.WithPlatform(&v1.Platform{
			OS:           "linux",
			Architecture: arch,
		}))
		if err != nil {
			log.Fatalf("pull %s: %v", arch, err)
		}
		out := fmt.Sprintf("internal/alpine_%s.tar", arch)
		if err := crane.Save(img, saveRef, out); err != nil {
			log.Fatalf("save %s: %v", arch, err)
		}
		log.Printf("wrote %s", out)
	}
}
