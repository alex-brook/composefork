package internal

import _ "embed"

//go:embed system_arm64.tar
var systemImageTarball []byte
