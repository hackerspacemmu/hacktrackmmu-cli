package version

import (
	"fmt"
)

var (
	Version = "dev"
)

type Info struct {
	Version string `json:"version"`
}

func GetInfo() Info {
	return Info{
		Version: Version,
	}
}

func (i Info) String() string {
	return fmt.Sprintf("ht %s", i.Version)
}
