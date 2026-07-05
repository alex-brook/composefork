package internal

import _ "embed"

//go:embed debian_arm64.tar
var systemImageTarball []byte
