package main

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"
)

func execute(argSegments *[]string, argTimeout *time.Duration) {
	var cmds []string

	segmentOutput := struct {
		sync.RWMutex
		segments map[string]Segment
	}{segments: make(map[string]Segment)}

	wg := new(sync.WaitGroup)

	re := regexp.MustCompile("\\s*=\\s*")
	for _, argSegment := range *argSegments {
		split := re.Split(argSegment, 2)
		cmd := split[0]
		hasParam := len(split) == 2

		param := ""
		if hasParam {
			param = split[1]
			if _, ok := segmentFuncsWithParam[cmd]; !ok {
				println("unknown segment (parameter needed):", cmd)
				os.Exit(1)
			}
		} else {
			if _, ok := segmentFuncsWithoutParams[cmd]; !ok {
				println("unknown segment:", cmd)
				os.Exit(1)
			}
		}
		cmds = append(cmds, cmd)

		wg.Add(1)
		go func(argCmd string, hasParam bool, argParam string) {
			defer wg.Done()

			chanSegment := make(chan Segment, 1)
			go func() {
				seg := makeSegment()
				if hasParam {
					segmentFuncsWithParam[argCmd](&seg, argParam)
				} else {
					segmentFuncsWithoutParams[argCmd](&seg)
				}
				chanSegment <- seg
			}()

			select {
			case seg := <-chanSegment:
				segmentOutput.Lock()
				segmentOutput.segments[argCmd] = seg
				segmentOutput.Unlock()
			case <-time.After(*argTimeout):
				seg := makeSegment()
				seg.AddHeadline("Segment ", argCmd, " timed out")
				segmentOutput.Lock()
				segmentOutput.segments[argCmd] = seg
				segmentOutput.Unlock()
			}

		}(cmd, hasParam, param)
	}

	wg.Wait()

	// print headlines
	for _, cmd := range cmds {
		segment := segmentOutput.segments[cmd]
		for _, headline := range segment.headLines {
			fmt.Println(headline)
		}
	}

	isFirstPrinted := false
	lastNonEmptyColor := SegmentColor{-1, -1}

	// print segments
	for i, cmd := range cmds {
		segment := segmentOutput.segments[cmd]
		if segment.IsPrintable() {
			if !isFirstPrinted {
				isFirstPrinted = true
				fmt.Print(segment.beginColor.fmt())
				fmt.Print(" ")
			} else {
				sepColor := SegmentColor{fg: lastNonEmptyColor.bg, bg: segment.beginColor.bg}
				fmt.Print(" ")
				fmt.Print(sepColor.fmt())
				fmt.Print(IconSegmentSep)
				fmt.Print(" ")
			}
		}
		if i == len(cmds)-1 { // last
			currentNonEmptyColor := SegmentColor{-1, -1}
			if segment.currentColor.IsEmpty() {
				currentNonEmptyColor = lastNonEmptyColor
			} else {
				currentNonEmptyColor = segment.currentColor
			}
			segment.AddText(" ", resetColors(), fmtColorFg(currentNonEmptyColor.bg), IconSegmentSep, " ", resetColors())

		}
		fmt.Print(segment.text)
		if !segment.currentColor.IsEmpty() {
			lastNonEmptyColor = segment.currentColor
		}
	}
}
