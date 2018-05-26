package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

var (
	from = flag.String("from", "", "file to copy")
	to   = flag.String("to", "", "file to copy to (default: <from>.bkp)")
)

func init() {
	log.SetFlags(0)
}

func copyFile(to, from string) error {
	log.Printf("Copying contents of %s to %s...", from, to)

	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(to)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func main() {
	if err := func() error {
		flag.Parse()

		if *from == "" {
			return errors.New("`from` must be specified")
		}

		if *to == "" {
			*to = fmt.Sprintf("%s.bkp", *from)
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return errors.Wrap(err, "failed to create a new watcher")
		}
		defer watcher.Close()

		if err = watcher.Add(*from); err != nil {
			return errors.Wrapf(err, "failed to watch %s", *from)
		}

		for {
			select {
			case err := <-watcher.Errors:
				return errors.Wrapf(err, "error while watching %s", *from)
			case event := <-watcher.Events:
				switch {
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					for {
						if err := copyFile(*from, *to); err != nil {
							log.Printf("Failed to restore %s: %s. Retrying in 1 second...", *from, err)
							time.Sleep(time.Second)
						}
						break
					}

					if err := watcher.Add(*from); err != nil {
						return errors.Wrapf(err, "failed to re-establish watcher for %s", *from)
					}

				case event.Op&fsnotify.Write != fsnotify.Write:
					continue
				}

				if err := copyFile(*to, *from); err != nil {
					return errors.Wrapf(err, "failed to backup %s", *from)
				}
			}
		}
	}(); err != nil {
		log.Fatal(err)
	}
}
