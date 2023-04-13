package main

import (
	_ "xyhelper-goframe/internal/packed"

	_ "github.com/cool-team-official/cool-admin-go/contrib/drivers/sqlite"

	_ "xyhelper-goframe/modules"

	"github.com/gogf/gf/v2/os/gctx"

	"xyhelper-goframe/internal/cmd"
)

func main() {
	// gres.Dump()
	cmd.Main.Run(gctx.New())
}
