package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/olekukonko/tablewriter"
)

func statesRunning(al map[string]*nomad.TaskState) (bool, error) {
	if len(al) == 0 {
		return false, nil
	}

	for task, allocState := range al {
		switch allocState.State {
		case structs.TaskStateDead:
			return false, errors.New(fmt.Sprintf("task %s is %s", task, structs.TaskStateDead))
		case structs.TaskStatePending:
			return false, nil
		case structs.TaskStateRunning:
			continue
		default:
			return false, nil
		}
	}

	return true, nil
}

func servicesPassing(checks consul.HealthChecks) bool {
	for _, ch := range checks {
		switch ch.Status {
		case consul.HealthPassing:
			continue
		default:
			return false
		}
	}

	return true

}

// Limits the length of the string.
func limit(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length]
}

func printReport(nCl *nomad.Client, cCl *consul.Client, id string) {

	allocs, _, err := nCl.Evaluations().Allocations(id, nil)

	if err != nil {
		panic(err)
	}

	for _, a := range allocs {
		alloc, _, err := nCl.Allocations().Info(a.ID, nil)

		if err != nil {
			panic(err)
		}

		table := tablewriter.NewWriter(os.Stdout)

		header := []string{"Job", "Alloc-ID", "Task", "State", "Failed", "Restarts", "Events"}
		table.SetHeader(header)

		for task, tstate := range alloc.TaskStates {

			events := strings.Join(formatTaskStatus(tstate), "")

			var col string
			switch tstate.State {
			case structs.TaskStateDead:
				col = "\033[0;31m%s\033[0m"
			case structs.TaskStatePending:
				col = "\033[1;33m%s\033[0m"
			case structs.TaskStateRunning:
				col = "\033[0;32m%s\033[0m"
			}

			state := fmt.Sprintf(col, tstate.State)

			data := []string{*alloc.Job.Name, limit(alloc.ID, 8), task, state, strconv.FormatBool(tstate.Failed), strconv.FormatUint(tstate.Restarts, 10), events}

			table.Append(data)
		}

		table.Render()

		table2 := tablewriter.NewWriter(os.Stdout)
		header2 := []string{"Service", "Status"}
		table2.SetHeader(header2)

		for _, tg := range alloc.Job.TaskGroups {

			for _, tk := range tg.Tasks {

				for _, ts := range tk.Services {

					if len(ts.Checks) != 0 {

						checks, _, err := cCl.Health().Checks(ts.Name, nil)

						if err != nil {
							panic(err)
						}

						for _, ch := range checks {

							var col string
							switch ch.Status {
							case consul.HealthCritical:
								col = "\033[0;31m%s\033[0m"
							case consul.HealthWarning:
								col = "\033[1;33m%s\033[0m"
							case consul.HealthPassing:
								col = "\033[0;32m%s\033[0m"
							default:
								col = "\033[1;30m%s\033[0m"
							}

							status := fmt.Sprintf(col, ch.Status)
							data := []string{ch.ServiceName, status}

							table2.Append(data)
						}

					}
				}
			}
		}

		table2.Render()

	}

}

func main() {

	evalID := flag.String("id", "", "evaluation id (Required)")
	timeout := flag.Duration("timeout", 10*time.Second, "how long to wait before assuming it failed")

	flag.Parse()

	if *evalID == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := nomad.DefaultConfig()

	nomadClient, err := nomad.NewClient(config)

	if err != nil {
		panic(err)
	}

	consulClient, err := consul.NewClient(consul.DefaultConfig())

	if err != nil {
		panic(err)
	}

	allocs, _, err := nomadClient.Evaluations().Allocations(*evalID, nil)

	if err != nil {
		panic(err)
	}

	fmt.Printf("evaluation \033[1;33m%s\033[0m has \033[1;32m%d\033[0m allocations\n", *evalID, len(allocs))

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)

	defer cancel()

	var wg sync.WaitGroup

	wg.Add(len(allocs))

	quit := make(chan int)

	done := make(chan int)

	for _, al := range allocs {

		go func(c context.Context, a *nomad.AllocationListStub) {
			defer wg.Done()

			readyChan := make(chan int)

			go func() {

				for {
					alloc, _, err := nomadClient.Allocations().Info(a.ID, nil)

					if err != nil {
						panic(err)
					}

					ready, err := statesRunning(alloc.TaskStates)

					if err != nil {
						fmt.Printf("\033[0;31mAllocation %s error:\033[0m %s\n", alloc.ID, err.Error())
						quit <- 1
					}

					if ready {
						readyChan <- 0
						return
					}
				}

			}()

			select {
			case <-c.Done():
				fmt.Printf("\033[0;31mAllocation timed up:\033[0m %s\n", a.ID)
				return
			case <-readyChan:
				fmt.Printf("\033[0;32mAllocation running:\033[0m %s\n", a.ID)
			}

			alloc, _, err := nomadClient.Allocations().Info(a.ID, nil)

			if err != nil {
				panic(err)
			}

			var sg sync.WaitGroup

			for _, tg := range alloc.Job.TaskGroups {

				for _, tk := range tg.Tasks {

					for _, ts := range tk.Services {

						if len(ts.Checks) != 0 {
							sg.Add(1)

							go func(s *nomad.Service) {
								defer sg.Done()
								for {
									checks, _, err := consulClient.Health().Checks(s.Name, nil)

									if err != nil {
										panic(err)
									}

									passing := servicesPassing(checks)

									if passing {
										return
									}
								}
							}(ts)
						}

					}

				}

			}

			sg.Wait()

		}(ctx, al)
	}

	go func() {
		wg.Wait()
		done <- 0
	}()

	select {
	case code := <-quit:
		cancel()
		printReport(nomadClient, consulClient, *evalID)
		os.Exit(code)
	case <-done:
		cancel()
		printReport(nomadClient, consulClient, *evalID)
		fmt.Printf("\033[0;32mDone!\033[0m All allocations for evaluation %s are running\n", *evalID)
	case <-ctx.Done():
		printReport(nomadClient, consulClient, *evalID)
		fmt.Printf("\033[0;31mTimeout!\033[0m Allocations did not finish running within deadline of \033[0;31m%f\033[0m seconds\n", (*timeout).Seconds())
		os.Exit(1)
	}

}
