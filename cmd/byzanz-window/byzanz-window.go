package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	flag "github.com/ogier/pflag"
	"github.com/syohex/byzanz-window"
)

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

func getSelectedRectangle() (*byzanzArg, error) {
	rect, err := byzanz.SelectWindow()
	if err != nil {
		return nil, err
	}

	arg := &byzanzArg{
		x:      rect.X,
		y:      rect.Y,
		width:  rect.Width,
		height: rect.Height,
	}

	return arg, nil
}

var rePosition = regexp.MustCompile(`\s*Position: (\d+),(\d+)`)
var reGeometry = regexp.MustCompile(`\s*Geometry: (\d+)x(\d+)`)

func getWindowRectangle() (*byzanzArg, error) {
	fmt.Println("Select the window which you like to capture.")

	b, err := exec.Command(
		"xdotool",
		"selectwindow",
		"windowactivate",
		"getwindowgeometry").Output()
	if err != nil {
		return nil, err
	}
	s := string(b)

	var x, y, w, h int

	m := rePosition.FindAllStringSubmatch(s, -1)
	if m == nil {
		return nil, fmt.Errorf(`can't find Position: %v`, s)
	}
	x, err = strconv.Atoi(string(m[0][1]))
	if err != nil {
		return nil, fmt.Errorf(`can't find Position x: %v`, s)
	}
	y, err = strconv.Atoi(string(m[0][2]))
	if err != nil {
		return nil, fmt.Errorf(`can't find Position y: %v`, s)
	}

	m = reGeometry.FindAllStringSubmatch(s, -1)
	if m == nil {
		return nil, fmt.Errorf(`can't find Geometry: %v`, s)
	}
	w, err = strconv.Atoi(string(m[0][1]))
	if err != nil {
		return nil, fmt.Errorf(`can't find Geometry width: %v`, b)
	}
	h, err = strconv.Atoi(string(m[0][2]))
	if err != nil {
		return nil, fmt.Errorf(`can't find Geometry height: %v`, b)
	}

	arg := &byzanzArg{
		x:      x,
		y:      y,
		width:  w,
		height: h,
	}

	return arg, nil
}

func main() {
	duration := flag.IntP("duration", "d", 10, "Capture duration(second)")
	delay := flag.IntP("delay", "", 1, "Delay before recording(second)")
	cursor := flag.BoolP("cursor", "c", false, "Record mouse cursor")
	audio := flag.BoolP("audio", "a", false, "Record audio")
	rect := flag.BoolP("rectangle", "r", false, "Record specified rectangle")
	flag.Parse()

	outputGif := flag.Arg(0)
	if outputGif == "" {
		fmt.Printf("Please specified output GIF filename\n")
		os.Exit(1)
	}

	var arg *byzanzArg
	var err error
	if *rect {
		arg, err = getSelectedRectangle()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		arg, err = getWindowRectangle()
		if err != nil {
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
