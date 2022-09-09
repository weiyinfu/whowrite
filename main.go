package main

import (
	"errors"
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

var ma sync.Map //存储用户名到行数的映射
var running sync.WaitGroup
var enableParallel bool

func init() {
	enableParallel = true
	running = sync.WaitGroup{}
	ma = sync.Map{}
}
func shouldEnterDir(p string) bool {
	name := filepath.Base(p)
	if len(name) > 0 && name[0] == '.' {
		return false
	}
	if name == "node_modules" {
		return false
	}
	return true
}
func shouldEnterFile(p string) bool {
	isCode := false
	name := filepath.Base(p)
	for _, lang := range strings.Fields(".py .go .java .js .cpp .c .h .hpp .cs") {
		if strings.HasSuffix(name, lang) {
			isCode = true
			break
		}
	}
	if !isCode {
		return false
	}
	return true
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
		log.Println("reading dir error", p, err)
		return
	}
	for _, son := range sons {
		sonPath := path.Join(p, son.Name())
		if son.IsDir() {
			if !shouldEnterDir(sonPath) {
				continue
			}
			if enableParallel {
				running.Add(1)
				go handleDir(sonPath)
			} else {
				handleDir(sonPath)
			}
			continue
		}
		if !shouldEnterFile(sonPath) {
			continue
		}
		handleFile(sonPath)
	}
}
func runCommand(name string, args []string, folder string) (*string, *string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = folder
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("处理stdout错误")
		return nil, nil, err
	}
	if stdout == nil {
		log.Println("stderr为空")
		return nil, nil, errors.New("stdout is nil")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("处理stderr错误")
		return nil, nil, err
	}
	if stderr == nil {
		log.Println("stderr为空")
		return nil, nil, errors.New("stderr is nil")
	}
	defer stdout.Close()
	defer stderr.Close()
	err = cmd.Start()
	if err != nil {
		log.Println("start command error")
		return nil, nil, err
	}
	res, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Println("read all content error", err)
		return nil, nil, err
	}
	errorInfo, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Println("read stderr error", err)
		return nil, nil, err
	}
	content := string(res)
	errorContent := string(errorInfo)
	return &content, &errorContent, nil
}
func handleFile(p string) {
	content, errorContent, err := runCommand("git", []string{"blame", filepath.Base(p)}, filepath.Dir(p))
	if err != nil {
		log.Println("执行命令失败", err)
		return
	}
	if len(*errorContent) != 0 {
		if strings.Index(*errorContent, "no such path") != -1 {
			//如果不在git里面直接continue
			return
		}
		log.Printf("handle file %v error %v", p, *errorContent)
		return
	}
	if len(*content) == 0 {
		//文件可能没有内容
		return
	}
	processContent(*content)
}

func processContent(content string) {
	x, err := regexp.Compile("\\(.+?\\)")
	if err != nil {
		panic("compile regex error")
	}
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
		run()
		show()
	})
}
