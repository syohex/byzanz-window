// +build linux

package byzanz

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/cursorfont.h>

struct rectangle {
	int x;
	int y;
	int width;
	int height;
};

static int select_region(struct rectangle *rect)
{
	Display *dpy = XOpenDisplay(NULL);
	Window root = DefaultRootWindow(dpy);
	XEvent ev;

	GC sel_gc;
	XGCValues sel_gv;

	int done = 0, btn_pressed = 0;
	int x = 0, y = 0;
	unsigned int width = 0, height = 0;
	int start_x = 0, start_y = 0;

	Cursor cursor = XCreateFontCursor(dpy, XC_crosshair);

	XGrabPointer(dpy, root, True, (PointerMotionMask|ButtonPressMask|ButtonReleaseMask),
		     GrabModeAsync, GrabModeAsync, None, cursor, CurrentTime);

	sel_gv.function = GXinvert;
	sel_gv.subwindow_mode = IncludeInferiors;
	sel_gv.line_width = 1;
	sel_gc = XCreateGC(dpy, root, (GCFunction|GCSubwindowMode|GCLineWidth), &sel_gv);

	for (;;) {
		XNextEvent(dpy, &ev);
		switch (ev.type) {
		case ButtonPress:
			btn_pressed = 1;
			x = start_x = ev.xbutton.x_root;
			y = start_y = ev.xbutton.y_root;
			width = height = 0;
			break;
		case MotionNotify:
			if (btn_pressed) {
				XDrawRectangle(dpy, root, sel_gc, x, y, width, height);

				x = ev.xbutton.x_root;
				y = ev.xbutton.y_root;

				if (x > start_x) {
					width = x - start_x;
					x = start_x;
				} else {
					width = start_x - x;
				}
				if (y > start_y) {
					height = y - start_y;
					y = start_y;
				} else {
					height = start_y - y;
				}

				XDrawRectangle(dpy, root, sel_gc, x, y, width, height);
				XFlush(dpy);
			}
			break;
		case ButtonRelease:
			done = 1;
			break;
		default:
			break;
		}
		if (done)
			break;
	}

	XDrawRectangle(dpy, root, sel_gc, x, y, width, height);
	XFlush(dpy);

	XUngrabPointer(dpy, CurrentTime);
	XFreeCursor(dpy, cursor);
	XFreeGC(dpy, sel_gc);
	XSync(dpy, 1);

	rect->x = x;
	rect->y = y;
	rect->width = width;
	rect->height = height;

	return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type Rectangle struct {
	Width  int
	Height int
	X      int
	Y      int
}

func SelectWindow() (*Rectangle, error) {
	var tmp C.struct_rectangle

	ret := int(C.select_region((*C.struct_rectangle)(unsafe.Pointer(&tmp))))
	if ret != 0 {
		return nil, fmt.Errorf("Failed: selecting window")
	}

	rect := &Rectangle{
		Width:  int(tmp.width),
		Height: int(tmp.height),
		X:      int(tmp.x),
		Y:      int(tmp.y),
	}

	return rect, nil
}
