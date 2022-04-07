package main

import (
	"flag"
	"github.com/onionf91/eth-explorer/pkg/service"
)

func main() {

	explorer := service.NewExplorerService()
	startPtr := flag.Uint64("start", 0, "block number that scan process starting from")
	parallelsPtr := flag.Int("parallels", 0, "number of parallel processes")
	migratePtr := flag.Bool("migrate", false, "auto migrate database schema")

	flag.Parse()

	if *migratePtr {
		explorer.AutoMigrateSchema()
	}
	explorer.ScanBlockFrom(*startPtr, *parallelsPtr)
}
