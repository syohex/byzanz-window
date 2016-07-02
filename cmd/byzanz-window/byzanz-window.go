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
	"github.com/syohex/byzanz-window"
)

const VERSION = "0.02"

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

	if string(xproperty) == "" {
		// Fallback: On some platform(LXDE), 'xprop -id ID' returns nothing.
		// Then get window information by xdotool.
		// Window information by xdotool is missaligned on some platform(Xfce4).
		return getWindowRectangle(winidStr)
	}

	match = posRe.FindStringSubmatch(string(xproperty))
	if match == nil {
		// Some windows managers, such as i3, don't support _NET_FRAME_EXTENTS yet.
		// Ignore window frame information for such window managers(issue #5) and not
		// capture window frame.
		match = []string{"0", "0", "0", "0", "0"}
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

var rePosition = regexp.MustCompile(`\s*Position: (\d+),(\d+)`)
var reGeometry = regexp.MustCompile(`\s*Geometry: (\d+)x(\d+)`)

func getWindowRectangle(winidStr string) (*byzanzArg, error) {
	b, err := exec.Command("xdotool", "getwindowgeometry", winidStr).Output()
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
	rect, err := byzanz.SelectRectangle()
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

func main() {
	duration := flag.IntP("duration", "d", 10, "Capture duration(second)")
	delay := flag.IntP("delay", "", 1, "Delay before recording(second)")
	cursor := flag.BoolP("cursor", "c", false, "Record mouse cursor")
	audio := flag.BoolP("audio", "a", false, "Record audio")
	rect := flag.BoolP("rectangle", "r", false, "Record specified rectangle")
	version := flag.BoolP("version", "v", false, "Show version")
	flag.Parse()

	if *version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

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
