package main

import (
	"bufio"
	"code.google.com/p/go.exp/fsnotify"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func compileRegex(regexStr []string) []*regexp.Regexp {
	var regexes []*regexp.Regexp

	for _, val := range regexStr {
		r, err := regexp.Compile(val)
		check(err)
		regexes = append(regexes, r)
	}

	return regexes
}

func lineReader(config Configuration, filename string, regexes []*regexp.Regexp, continueRead chan bool) {
	mail := Mail{Name: config.Mail.Name, Email: config.Mail.Email, Password: config.Mail.Password, To: config.Mail.To}

	byteRead := int64(0)

	for {
		f, err := os.Open(filename)
		check(err)

		byteCount := int64(0)

		if byteRead > 0 {
			// if not our first read, skip to the last read pos
			_, err := f.Seek(byteRead, 0)
			check(err)

			byteCount = byteRead
		} else {
			// if first read, (virtually) seek to end of file
			info, err := f.Stat()
			check(err)

			byteRead = info.Size() + 1

			<-continueRead
			continue
		}

		// TODO: create our own reader instead of relying on ReadLine()
		bf := bufio.NewReader(f)

		for {
			// reader.ReadLine does a buffered read up to a line terminator,
			// handles either /n or /r/n, and returns just the line without
			// the /r or /r/n.
			line, isPrefix, err := bf.ReadLine()

			if err == io.EOF {
				// if end of file, break loop and wait for continueRead
				byteRead = byteCount
				break
			}

			check(err)

			// TODO: need to handle uber long lines
			if isPrefix {
				// skip line if too long to be read (4k is reasonable)
				continue
			}

			// +1 is accounting for \n or \r (does not account for \r\n)
			byteCount += int64(len(line) + 1)

			lineStr := string(line)

			matched := false

			// match with declared regex
			for _, val := range regexes {
				if val.MatchString(lineStr) {
					fmt.Println("matched regex:")
					fmt.Println("\t" + lineStr)
					matched = true
				}
			}

			if matched {
				hostname, _ := os.Hostname()

				lines := "From: " + config.Mail.Email + "\r\n"
				lines += "To: " + config.Mail.To + "\r\n"
				lines += "Subject: [" + hostname + "] Detected Error in " + filename + "\r\n"
				lines += "\r\n"

				lines += "===============THIS IS AN AUTOMATED EMAIL BY CYCLOPS===============\n\n"
				lines += "Log Excerpt:\n\n"
				lines += "1: " + lineStr + "\n"

				for i := 0; i < config.Excerpt_size; i++ {
					line, isPrefix, err := bf.ReadLine()

					if err == io.EOF {
						break
					}

					check(err)

					if isPrefix {
						continue
					}

					byteCount += int64(len(line) + 1)

					lines += strconv.Itoa(i+2) + ": " + string(line) + "\n"
				}

				lines += "\n\n===============THIS IS AN AUTOMATED EMAIL BY CYCLOPS===============\n"

				go mail.send(lines)
			}
		}

		// look ahead, if reaching EOF, wait for continueRead
		b, err := bf.Peek(1)
		_ = b

		if err == io.EOF {
			<-continueRead
		}
	}
}

// TODO: File existence validation, Regex array == File array length validation
func main() {
	var config Configuration
	config.parseConfig()

	var watchers []*fsnotify.Watcher

	done := make(chan bool)

	for index, filename := range config.Files {
		watcher, err := fsnotify.NewWatcher()
		check(err)

		watchers = append(watchers, watcher)

		continueRead := make(chan bool)

		go func() {
			for {
				select {
				case <-watcher.Event:
					//log.Println("event:", ev)
					time.Sleep(15 * time.Second)
					continueRead <- true
				case err := <-watcher.Error:
					log.Println("error:", err)
				}
			}
		}()

		regexes := compileRegex(config.Detect[index])

		go lineReader(config, filename, regexes, continueRead)

		err = watcher.Watch(filename)
		check(err)

		fmt.Println("watching: " + filename)
	}

	<-done

	for _, watcher := range watchers {
		watcher.Close()
	}
}
