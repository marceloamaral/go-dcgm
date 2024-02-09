package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
)

const (
	header = `# gpu   pwr  temp    sm   mem   enc   dec  mclk  pclk
# Idx     W     C     %     %     %     %   MHz   MHz`
)

var (
	connectAddr = flag.String("connect", "localhost", "Provide nv-hostengine connection address.")
	isSocket    = flag.String("socket", "0", "Connecting to Unix socket?")
)

// modelled on nvidia-smi dmon
// dcgmi dmon -e 155,150,203,204,206,207,100,101
func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flag.Parse()
	var cleanup func()
	var err error
	cleanup, err = dcgm.Init(dcgm.Standalone, *connectAddr, *isSocket)
	// c
	if err != nil {
		fmt.Printf("There is no DCGM daemon running in the host, starting the Standalone mode: %s", err)
		cleanup, err = dcgm.Init(dcgm.Embedded)
		if err != nil {
			fmt.Printf("Could not start DCGM. Error: %s", err)
		}
	}
	defer cleanup()

	gpus, err := dcgm.GetSupportedDevices()
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println("GetSupportedDevices", gpus)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	fmt.Println(header)
	for {
		select {
		case <-ticker.C:
			for _, gpu := range gpus {
				// groupName := fmt.Sprintf("devStatus%d", 42)
				// group, err := dcgm.NewDefaultGroup(groupName)
				// if err != nil {
				// 	return
				// }
				// fmt.Println("group", group, err)
				// // runinng ds
				// cmd := exec.Command("/usr/bin/dcgmi", "group", "-l")
				// // cmd := exec.Command("/usr/bin/dcgmi", "group", "-l", "-j")
				// stdout, err := cmd.StdoutPipe()
				// if err != nil {
				// 	log.Fatal(err)
				// }
				// if err := cmd.Start(); err != nil {
				// 	log.Fatal(err)
				// }
				// r = bufio.NewReader(stdout)
				// fmt.Println("Read:")
				// for {
				// 	b, err := r.ReadBytes('\n')
				// 	if err != nil {
				// 		break
				// 	}
				// 	fmt.Println(string(b))
				// }
				// os.Exit(0)

				st, err := dcgm.GetDeviceStatus(gpu)
				if err != nil {
					// log.Panicln(err)
					fmt.Println(gpu, err)
					continue
				}
				fmt.Printf("%5d %5d %5d %.3f %.3f %.3f %.3f %5d %5d\n",
					gpu, int64(st.Power), st.Temperature, st.Utilization.GPU, st.Utilization.Memory,
					st.Utilization.Encoder, st.Utilization.Decoder, st.Clocks.Memory, st.Clocks.Cores)
			}

		case <-sigs:
			return
		}
	}
}
