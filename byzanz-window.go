package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	flag "github.com/ogier/pflag"
)

func selectWindow() (int, error) {
	fmt.Println("Select the window which you like to capture.")

	bytes, err := exec.Command(`xdotool`, "selectwindow").Output()
	if err != nil {
		return 0, err
	}

	winidStr := string(bytes)
	winidStr = strings.TrimRight(winidStr, "\n")

	winid, err := strconv.ParseInt(winidStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(winid), nil
}

type byzanzArg struct {
	x        int
	y        int
	width    int
	height   int
	duration int
	delay    int
	cursor   bool
	audio    bool
	output   string
}

var xRe = regexp.MustCompile(`Absolute upper-left X:\s*(\d+)`)
var yRe = regexp.MustCompile(`Absolute upper-left Y:\s*(\d+)`)
var widthRe = regexp.MustCompile(`Width:\s*(\d+)`)
var heightRe = regexp.MustCompile(`Height:\s*(\d+)`)
var posRe = regexp.MustCompile(`_NET_FRAME_EXTENTS\(CARDINAL\) = (\d+), (\d+), (\d+), (\d+)`)

func getWindowInformation(winid int) (*byzanzArg, error) {
	winidStr := strconv.Itoa(winid)

	bytes, err := exec.Command(`xwininfo`, `-id`, winidStr).Output()
	if err != nil {
		return nil, err
	}

	wininfo := string(bytes)
	wininfo = strings.TrimRight(wininfo, "\n")

	var match []string
	match = xRe.FindStringSubmatch(wininfo)
	if match == nil {
		return nil, errors.New(`can't find 'x' position`)
	}

	x, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil, err
	}

	match = yRe.FindStringSubmatch(wininfo)
	if match == nil {
		return nil, errors.New(`can't find 'y' position`)
	}

	y, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil, err
	}

	match = widthRe.FindStringSubmatch(wininfo)
	if match == nil {
		return nil, errors.New(`can't find 'width'`)
	}

	width, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil, err
	}

	match = heightRe.FindStringSubmatch(wininfo)
	if match == nil {
		return nil, errors.New(`can't find 'height'`)
	}

	height, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil, err
	}

	xproperty, err := exec.Command(`xprop`, `-id`, winidStr).Output()
	if err != nil {
		return nil, err
	}

	match = posRe.FindStringSubmatch(string(xproperty))
	if match == nil {
		return nil, errors.New(`can't find 'position'`)
	}

	left, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil, err
	}

	right, err := strconv.ParseInt(match[2], 10, 32)
	if err != nil {
		return nil, err
	}

	top, err := strconv.ParseInt(match[3], 10, 32)
	if err != nil {
		return nil, err
	}

	bottom, err := strconv.ParseInt(match[4], 10, 32)
	if err != nil {
		return nil, err
	}

	arg := &byzanzArg{
		x:      int(x - left),
		y:      int(y - top),
		width:  int(width + left + right),
		height: int(height + top + bottom),
	}

	return arg, nil
}

func focusWindow(winid int) error {
	fmt.Println("Press enter when you are ready to capture.")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		_ = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	err := exec.Command(`xdotool`, `windowactivate`, strconv.Itoa(winid)).Run()
	if err != nil {
		return err
	}
	return nil
}

func record(arg *byzanzArg) error {
	var cmdArgs []string
	if arg.cursor {
		cmdArgs = append(cmdArgs, `-c`)
	}

	if arg.audio {
		cmdArgs = append(cmdArgs, `-a`)
	}

	cmdArgs = append(cmdArgs,
		`-x`, strconv.Itoa(arg.x),
		`-y`, strconv.Itoa(arg.y),
		`-w`, strconv.Itoa(arg.width),
		`-h`, strconv.Itoa(arg.height),
		`-d`, strconv.Itoa(arg.duration),
		`--delay`, strconv.Itoa(arg.delay),
		arg.output,
	)

	cmd := exec.Command(`byzanz-record`, cmdArgs...)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

var xrectselRe = regexp.MustCompile(`(\d+)x(\d+)\+(\d+)\+(\d+)`)

func getSelectedRectangle() (*byzanzArg, error) {
	bytes, err := exec.Command(`xrectsel`).Output()
	if err != nil {
		return nil, err
	}

	rectangle := string(bytes)
	rectangle = strings.TrimRight(rectangle, "\n")

	match := xrectselRe.FindStringSubmatch(rectangle)
	if match == nil {
		return nil, errors.New(`Can't find 'rectangle' information` + rectangle)
	}

	x, err := strconv.ParseInt(match[3], 10, 32)
	if err != nil {
		return nil, err
	}

	y, err := strconv.ParseInt(match[4], 10, 32)
	if err != nil {
		return nil, err
	}

	width, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil, err
	}

	height, err := strconv.ParseInt(match[2], 10, 32)
	if err != nil {
		return nil, err
	}

	arg := &byzanzArg{
		x: int(x),
		y: int(y),
		width: int(width),
		height: int(height),
	}

	return arg, nil
}

func main() {
	duration := flag.IntP("duration", "d", 10, "Capture duration(duration)")
	delay := flag.IntP("delay", "", 1, "Delay before start")
	cursor := flag.BoolP("cursor", "c", false, "Record mouse cursor")
	audio := flag.BoolP("audio", "a", false, "Record audio")
	rect := flag.BoolP("rectangle", "r", false, "Record specified rectangleRecord audio")
	flag.Parse()

	outputGif := flag.Arg(0)
	if outputGif == "" {
		fmt.Printf("Please specified output GIF filename\n")
		os.Exit(1)
	}

	var arg *byzanzArg
	if *rect {
		var err error
		arg, err = getSelectedRectangle()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		winid, err := selectWindow()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		arg, err = getWindowInformation(winid)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := focusWindow(winid); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	arg.duration = *duration
	arg.delay = *delay
	arg.cursor = *cursor
	arg.audio = *audio
	arg.output = outputGif

	if err := record(arg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
