package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var ma sync.Map
var running sync.WaitGroup
var enableParallel bool

func init() {
	enableParallel = true
	running = sync.WaitGroup{}
	ma = sync.Map{}
}
func handleDir(p string) {
	if enableParallel {
		defer running.Done()
	}
	info, err := os.Stat(p)
	if err != nil {
		panic(fmt.Sprintf("get info error %v filepath=%v", err, p))
	}
	if !info.IsDir() {
		panic(fmt.Sprintf("%v is a file", p))
	}
	sons, err := ioutil.ReadDir(p)
	if err != nil {
		log.Println(err)
		return
	}
	for _, son := range sons {
		name := son.Name()
		if len(name) > 0 && name[0] == '.' {
			continue
		}
		if name == "node_modules" {
			continue
		}
		sonPath := path.Join(p, name)
		if son.IsDir() {
			running.Add(1)
			if enableParallel {
				go handleDir(sonPath)
			} else {
				handleDir(sonPath)
			}
			continue
		}
		isCode := false
		for _, lang := range strings.Fields(".py .go .java .js .cpp") {
			if strings.HasSuffix(name, lang) {
				isCode = true
				break
			}
		}
		if !isCode {
			continue
		}
		handleFile(sonPath)
	}
}
func handleFile(p string) {
	cmd := exec.Command("git", "blame", p)
	cmd.Dir = filepath.Dir(p)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("处理stdout错误")
		return
	}
	defer stdout.Close()
	err = cmd.Start()
	if err != nil {
		log.Println("start command error")
	}
	res, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Println("read all content error")
		return
	}
	content := string(res)
	processContent(content)
}

func processContent(content string) {
	x, err := regexp.Compile("\\(.+?\\)")
	if err != nil {
		panic("compile regex error")
	}
	//fmt.Println(content)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		res := x.FindString(line)
		res = res[1 : len(res)-1]
		fields := strings.Fields(res)
		name := fields[0]
		emit(name)
	}
}
func emit(name string) {
	count, ok := ma.Load(name)
	if !ok {
		count = 0
	}
	cnt := count.(int)
	ma.Store(name, cnt+1)
}
func show() {
	type node struct {
		name  string
		count int
	}
	var a []node
	ma.Range(func(name, count interface{}) bool {
		a = append(a, node{name: name.(string), count: count.(int)})
		return true
	})
	sort.Slice(a, func(i, j int) bool {
		return a[i].count > a[j].count
	})
	for _, no := range a {
		fmt.Println(no.name, no.count)
	}
}
func timeit(f func()) {
	beginTime := time.Now()
	f()
	endTime := time.Now()
	duration := endTime.Sub(beginTime)
	fmt.Printf("Time used %v", duration)
}
func run() {
	if enableParallel {
		running.Add(1)
		go handleDir(".")
		running.Wait()
	} else {
		handleDir(".")
	}
}
func main() {
	timeit(func() {
		//handleDir("xxxxbes")
		//handleFile("xxxxxx")
		run()
		show()
	})
}
