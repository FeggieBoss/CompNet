package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

type Router struct {
	Router string   `json:"router"`
	Edges  []string `json:"edges"`
}

type RouterTablePiece struct {
	source      string
	destinition string
	nextHop     string
	metric      int
}

func printTable(mutex *sync.Mutex, t []RouterTablePiece, header string) {
	mutex.Lock()
	fmt.Printf("%s of router %s table\n", header, t[0].source)
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', 0)
	fmt.Fprintln(w, "[Source IP]\t[Destination IP]\t[Next Hop]\t[Metric]")
	for _, piece := range t {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
			piece.source,
			piece.destinition,
			piece.nextHop,
			piece.metric,
		)
	}
	w.Flush()
	fmt.Println()
	mutex.Unlock()
}

const MAX_QUEUE = 500000
const TIME_LIMIT_SEC = 1

func main() {
	var fileName = flag.String("file", "conf.json", `Configuration file name`)
	flag.Parse()

	jsonFile, err := os.Open(*fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	data, _ := ioutil.ReadAll(jsonFile)
	var routers []Router
	err = json.Unmarshal(data, &routers)
	if err != nil {
		log.Fatalf("Wrong configuration file format: %v", err)
	}

	cs := make(map[string]chan []RouterTablePiece, len(routers))
	for _, router := range routers {
		cs[router.Router] = make(chan []RouterTablePiece, MAX_QUEUE)
	}

	ctx, cancel := context.WithTimeout(context.Background(), TIME_LIMIT_SEC*time.Second)
	defer cancel()

	fmt.Println("START")
	var (
		wg    = sync.WaitGroup{}
		mutex = sync.Mutex{}
	)
	for _, router := range routers {
		localcs := make([]chan []RouterTablePiece, len(router.Edges))
		for i, edge := range router.Edges {
			localcs[i] = cs[edge]
		}

		wg.Add(1)
		go func(router Router) {
			router.processRouter(ctx, &wg, &mutex, cs[router.Router], localcs)
		}(router)
	}

	wg.Wait()
}

func squeezeMap[K comparable, V any](m map[K]V) []V {
	var (
		r = make([]V, len(m))
		i = 0
	)
	for _, v := range m {
		r[i] = v
		i++
	}
	return r
}

func (r *Router) processRouter(ctx context.Context, wg *sync.WaitGroup, mutex *sync.Mutex, myc chan []RouterTablePiece, cs []chan []RouterTablePiece) {
	defer wg.Done()

	var (
		step = 0
		t    = make(map[string]RouterTablePiece, len(r.Edges))
		st   []RouterTablePiece
	)

	for _, edge := range r.Edges {
		t[edge] = RouterTablePiece{
			source:      r.Router,
			destinition: edge,
			nextHop:     edge,
			metric:      1,
		}
	}
	st = squeezeMap(t)
	printTable(mutex, st, fmt.Sprintf("Simulation step %d", step))
	for _, c := range cs {
		c <- st
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case updates := <-myc:
			isBetter := false
			for _, update := range updates {
				if update.destinition == r.Router {
					continue
				}
				if old, ok := t[update.destinition]; ok && old.metric <= update.metric+1 {
					continue
				}
				t[update.destinition] = RouterTablePiece{
					source:      r.Router,
					destinition: update.destinition,
					nextHop:     update.source,
					metric:      update.metric + 1,
				}
				isBetter = true
			}

			if isBetter {
				st = squeezeMap(t)
				for _, c := range cs {
					c <- st
				}
				step++
				printTable(mutex, st, fmt.Sprintf("Simulation step %d", step))
			}
		}
	}
	printTable(mutex, st, "Final step")
}
