package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli"
)

var (
	VERSION = "v0.0.0-dev"
)

type FsWatch struct {
	*fsnotify.Watcher
}

func NewFsWatcher() (*FsWatch, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FsWatch{
		Watcher: watcher,
	}, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "fswatch"
	app.Version = VERSION
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		return
	}
}

func run(c *cli.Context) {
	dir := c.Args().Get(0)
	if dir == "" {
		dir = "./"
	}

	fmt.Printf("Begin watching: %s\n", dir)

	var wg sync.WaitGroup

	wg.Add(1)

	watcher, err := NewFsWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer watcher.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					wg.Done()
					close(sigs)
					os.Exit(0)
				}

				if event.Op&fsnotify.Remove == fsnotify.Remove && event.Name != "" {
					fmt.Println("REMOVE:", event.Name)

					if fi, err := os.Lstat(event.Name); err == nil && fi.IsDir() {
						watcher.RecursiveRemove(event.Name)
					}
					break
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					fi, err := os.Lstat(event.Name)
					if err != nil {
						break
					}
					if strings.HasPrefix(fi.Name(), ".") {
						break
					}

					fmt.Println("CREATE:", event.Name)
					if fi.IsDir() {
						watcher.RecursiveAdd(event.Name)
					}
					break
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("WRITE:", event.Name)
					break
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					wg.Done()
					close(sigs)
					os.Exit(0)
				}
				log.Println("error:", err)
			}
		}
	}()

	go func() {
		<-sigs
		wg.Done()
	}()

	err = watcher.Add(dir)
	if err != nil {
		fmt.Println("[Error!] Can't watch the root directory.")
		os.Exit(0)
	}

	watcher.RecursiveAdd(dir)

	wg.Wait()

	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
}

func (w *FsWatch) RecursiveAdd(dir string) error {
	return filepath.Walk(dir, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), ".") {
				return nil
			}
			if err = w.Add(walkPath); err != nil {
				return err
			}
		}
		return nil
	})
}

func (w *FsWatch) RecursiveRemove(dir string) error {
	return filepath.Walk(dir, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), ".") {
				return nil
			}
			if err = w.Remove(walkPath); err != nil {
				return err
			}
		}
		return nil
	})
}
