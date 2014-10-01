package linklist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// LinkList abstracts a simple map
type LinkList struct {
	links map[string]bool
}

const seenUrlsFile = "seenUrls"

var llist = LinkList{make(map[string]bool)}

// Save json marshals the current map into the given filename
// WARNING: no check is done of existing file, so be careful
func (l LinkList) Save(filename string) {
	data, err := json.Marshal(l.links)
	if err != nil {
		fmt.Println("Cannot save seen url list")
		panic(err)
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		panic(err)
	}
}

func saveMonitor() {
	sigc := make(chan os.Signal, 1)
	tickc := time.NewTicker(time.Second * 3600).C
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	var c os.Signal
	for {
		select {
		case c = <-sigc:
			fmt.Println("Signal received, saving seen urls")
		case <-tickc:
			fmt.Println("Timeout elapsed, saving seen urls")
		}
		llist.Save(seenUrlsFile)
		if c != syscall.SIGHUP && c != nil {
			os.Exit(0)
		}
	}
}

func init() {
	if _, err := os.Stat(seenUrlsFile); os.IsNotExist(err) {
		fmt.Printf("File %s not found, assuming new one needed\n", seenUrlsFile)
	} else {
		data, err := ioutil.ReadFile(seenUrlsFile)
		if err != nil {
			fmt.Println("Error reading seen url file:", err)
			os.Exit(1)
		}
		err = json.Unmarshal(data, &llist.links)
		if err != nil {
			fmt.Println("Error demarshaling seen url file:", err)
			os.Exit(1)
		}
	}
	// start save monitor to save the seen url list each hour OR when a signal is received
	go saveMonitor()
}

// Uniq returns false if the link was already "seen" by a previous
// call or adds the link to the list and returns true
func Uniq(linkin string) (seen bool) {
	_, present := llist.links[linkin]
	if present {
		return false
	}
	fmt.Println(fmt.Sprintf("Added link %s", linkin))
	llist.links[linkin] = true
	return true
}
