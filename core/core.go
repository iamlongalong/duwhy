package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/reactivex/rxgo/v2"
)

var defers = []func(){}

type DUOption struct {
	BasePaths []string
}

type DULine struct {
	SizeKB       int
	ModifiedTime time.Time
	Paths        []string

	DUOption *DUOption
}

type PathNode struct {
	PathName     string
	SizeKB       int
	ModifiedTime time.Time

	Childs []PathNode
}

func xinit() rxgo.Producer {
	f, err := os.OpenFile("/Users/long/go/src/duwhy/xx.log", os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(f)
	defers = append(defers, func() { f.Close() })

	ch := make(chan rxgo.Item)

	go func() {
		defer func() { close(ch) }()

		for {
			l, _, err := r.ReadLine()
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("finish read line")
					return
				}

				log.Print("read line err", err)
				return
			}

			nl := make([]byte, len(l))
			copy(nl, l)

			duline, err := parseDuLine(nl)
			if err != nil {
				log.Print(err)
				continue
			}

			i := rxgo.Item{V: duline}
			i.SendBlocking(ch)
		}
	}()

	return func(ctx context.Context, next chan<- rxgo.Item) {
		defer func() {
			fmt.Println("stopped next")
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case x, ok := <-ch:
				if !ok {
					return
				}

				next <- x
			}
		}
	}
}

func runit() {
	p := rxgo.Create([]rxgo.Producer{xinit()})

	ch := p.Observe()

	// var lastNode DULine

	for item := range ch {
		fmt.Println(item.V)
		time.Sleep(time.Second)

		// last = item
	}

	log.Print("finish observe")
}

func parseDuLine(b []byte) (DULine, error) {
	l := DULine{}

	infos := bytes.Split(b, []byte("\t"))
	if len(infos) != 3 {
		return l, errors.Errorf("line block num not 3 : %s", string(b))
	}

	var err error
	l.SizeKB, err = strconv.Atoi(string(infos[0]))
	if err != nil {
		return l, errors.Wrapf(err, "parse size fail : %s", string(b))
	}

	l.ModifiedTime, err = time.Parse("2006-01-02 15:04", string(infos[1]))
	if err != nil {
		return l, errors.Wrapf(err, "parse modified time fail : %s", string(b))
	}

	ps := bytes.Split(infos[2], []byte("/"))
	l.Paths = make([]string, 0, len(ps))

	for _, p := range ps {
		l.Paths = append(l.Paths, string(p))
	}

	return l, nil
}
