package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	cb "github.com/atotto/clipboard"
	"github.com/fsnotify/fsnotify"
)

var (
	//DEFINE FLAG DEFAULTS
	filename = demoPath + "\\demo\\MP\\qconsole.log"
	numLines = 1

	sfilename = silverPath + "\\RPM\\qconsole.log"

	//ErrNoFilename is thrown when the path the the file to tail was not given
	ErrNoFilename = errors.New("you must provide the path to a file in the \"-file\" flag")

	//ErrInvalidLineCount is thrown when the user provided 0 (zero) as the value for number of lines to tail
	ErrInvalidLineCount = errors.New("you cannot tail zero lines")

	demoPath   = "C:\\Users\\bjorn\\Documents\\DEMO"
	silverPath = "C:\\Users\\bjorn\\Documents\\SILVER"
)

func startWatch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				log.Printf("%s %s\n", event.Name, event.Op)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}

	}()

	err = watcher.Add(demoPath + "\\demo\\MP")
	if err != nil {
		log.Fatal("Add failed:", err)
	}
	<-done
}

// GoTail IS THE FUNCTION THAT ACTUALLY DOES THE "TAILING"
// this can be used this package is imported into other golang projects
func GoTail(filename string, numLines int) ([]byte, error) {
	//MAKE SURE FILENAME IS GIVEN
	//actually, a path to the file
	if len(filename) == 0 {
		return nil, ErrNoFilename
	}

	//MAKE SURE USER WANTS TO GET AT LEAST ONE LINE
	if numLines == 0 {
		return nil, ErrInvalidLineCount
	}

	//OPEN FILE
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	//SEEK BACKWARD CHARACTER BY CHARACTER ADDING UP NEW LINES
	//offset must start at "-1" otherwise we are already at the EOF
	//"-1" from numLines since we ignore "last" newline in a file
	numNewLines := 0
	var offset int64 = -1
	var finalReadStartPos int64
	for numNewLines <= numLines-1 {
		//seek to new position in file
		startPos, err := file.Seek(offset, 2)
		if err != nil {
			return nil, err
		}

		//make sure start position can never be less than 0
		//aka, you cannot read from before the file starts
		if startPos == 0 {
			//set to -1 since we +1 to this below
			//the position will then start from the first character
			finalReadStartPos = -1
			break
		}

		//read the character at this position
		b := make([]byte, 1)
		_, err = file.ReadAt(b, startPos)
		if err != nil {
			return nil, err
		}

		//ignore if first character being read is a newline
		if offset == int64(-1) && string(b) == "\n" {
			offset--
			continue
		}

		//if the character is a newline
		//add this to the number of lines read
		//and remember position in case we have reached our target number of lines
		if string(b) == "\n" {
			numNewLines++
			finalReadStartPos = startPos
		}

		//decrease offset for reading next character
		//remember, we are reading backward!
		offset--
	}

	//READ TO END OF FILE
	//add "1" here to move offset from the newline position to first character in line of text
	//this position should be the first character in the "first" line of data we want
	b := make([]byte, 100)
	_, err = file.ReadAt(b, finalReadStartPos+1)
	if err == io.EOF {
		return bytes.Trim(b, "\x00"), nil
	} else if err != nil {
		return nil, err
	}

	//special case
	//if text is read, then err == io.EOF should hit
	//there should *never* not be an error above
	//so this line should never return
	return nil, nil
}

func ImpersonateLastChatter() {
	//TAIL
	b, err := GoTail(filename, numLines)
	if err != nil {
		fmt.Println(err)
	}

	// Time for Regexp
	text := string(b)
	re := regexp.MustCompile(`[^:]+`)
	line := re.FindAllString(text, -1)
	if len(line) == 2 {
		name := line[0]

		paddingAmount := 75
		padding := strings.Repeat(" ", paddingAmount)
		final := fmt.Sprintf("%s %s: ^2lol", padding, name)

		fmt.Println("Attempting to impersonate", name)
		//a := strings.ReplaceAll(string(b), "\n", "")
		//c := strings.TrimSpace(a)
		//fmt.Println(c)

		//Write last line to clipboard
		er := cb.WriteAll(final)
		if er != nil {
			fmt.Println(err)
		}
	}

}

func CopyTextLastChatter() {
	//TAIL
	b, err := GoTail(filename, numLines)
	if err != nil {
		fmt.Println(err)
	}

	// Time for Regexp

	text := string(b)
	re := regexp.MustCompile(`[^:]+`)
	line := re.FindAllString(text, -1)

	if len(line) == 2 {

		chat := line[1]

		er := cb.WriteAll(chat)
		if er != nil {
			fmt.Println(err)
		}
	}

}
