package internal

import _ "embed"

//go:embed debian_amd64.tar
var systemImageTarball []byte
