package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	operation := flag.String("op", "dump", "dump/restore")
	dumpFile := flag.String("dump-file", "", "Dump file path")
	//metricPtr := flag.String("metric", "chars", "Metric {chars|words|lines}.")
	//uniquePtr := flag.Bool("unique", false, "Measure unique values of a metric.")

	//fmt.Printf("Args list %v", flag.Args())
	//if len(flag.Args()) < 1 {
	//	fmt.Printf("Need to specify atleast dump/restore:")
	//	return
	//}
	//flag.Args()
	flag.Parse()

	fmt.Printf("operation is %s", *operation)
	//if *operation != "dump" && *operation != "restore" {
	//	fmt.Printf("Need to specify atleast dump/restore:")
	//	return
	//}
	etcdClient, err := GetEtcdClient()
	if err != nil {
		fmt.Printf("err creating etcd client %v", err)
		return
	}

	// Access parsed values
	fmt.Printf("etcdbackuprestore cli, mode: %s, dumpFile: %s\n", flag.Arg(0), *dumpFile)

	if *operation == "dump" {
		fmt.Printf("op is dump")
		Backup(etcdClient, *dumpFile)
	} else if *operation == "restore" {
		fmt.Printf("op is restore")

		fileData, err := os.ReadFile(*dumpFile)
		if err != nil {
			fmt.Printf("err reading dump json file  %v", err)
			return
		}

		dataMap, err := parseEtcdJSON(string(fileData))
		if err != nil {
			fmt.Printf("err parsing dump json file  %v", err)
			return
		}
		Restore(etcdClient, dataMap)
	} else {
		fmt.Printf("op %s is not supported", *operation)
	}

}
