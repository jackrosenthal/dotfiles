package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	stddraw "image/draw"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackrosenthal/dotfiles/gobin/waybar-calendar-popup/wlrlayershell"
	"github.com/pdf/go-wayland/client"
	_jsii_viewporter "github.com/pdf/go-wayland/stable/viewporter"
	fractional_scale "github.com/pdf/go-wayland/staging/fraction-sclae-v1"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	pidFile             = "/tmp/waybar-calendar-popup.pid"
	moduleTimeFmt       = "Mon Jan 02  15:04:05"
	monthTitleWidth     = 20
	monthGapChars       = 2
	popupPaddingLeft    = 16
	popupPaddingRight   = 28
	popupPaddingY       = 16
	popupBorderWidth    = 1
	popupBottomMargin   = 0
	popupRightMargin    = 8
	popupExclusiveZone  = 0
	layerNamespace      = "waybar-calendar-popup"
	layerSurfaceVersion = 4
	popupFontPath       = "/usr/share/fonts/TTF/JetBrainsMonoNerdFont-Regular.ttf"
	popupFontSize       = 15
	popupFontDPI        = 72
)

var (
	weekdayNames = []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

	colorBackground = color.RGBA{0x2a, 0x1f, 0x35, 0xff}
	colorBorder     = color.RGBA{0x4a, 0x42, 0x53, 0xff}
	colorText       = color.RGBA{0xf7, 0xf7, 0xf7, 0xff}
	colorMuted      = color.RGBA{0xc8, 0xc8, 0xc8, 0xff}
	colorHighlight  = color.RGBA{0xf7, 0xf7, 0xf7, 0xff}
	colorHighlightF = color.RGBA{0x11, 0x11, 0x11, 0xff}
)

type moduleOutput struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip,omitempty"`
}

type calendarLayout struct {
	face            font.Face
	scale           float64
	bufferWidth     int
	bufferHeight    int
	charWidth       int
	lineHeight      int
	ascent          int
	descent         int
	width           int
	height          int
	contentWidth    int
	contentHeight   int
	monthWidth      int
	monthGap        int
	padding         int
	borderWidth     int
	titleY          int
	weekdayY        int
	weeksTopY       int
	month1X         int
	month2X         int
	month3X         int
	dayCellWidth    int
	dayCellHeight   int
	highlightInsetX int
	highlightInsetY int
}

type popupApp struct {
	display      *client.Display
	registry     *client.Registry
	compositor   *client.Compositor
	shm          *client.Shm
	seat         *client.Seat
	pointer      *client.Pointer
	output       *client.Output
	outputScale  int32
	viewporter   *_jsii_viewporter.Viewporter
	viewport     *_jsii_viewporter.Viewport
	fractional   *fractional_scale.FractionalScaleManager
	surfaceScale *fractional_scale.FractionalScale
	preferred120 uint32
	layerShell   *wlrlayershell.LayerShell
	surface      *client.Surface
	layerSurface *wlrlayershell.LayerSurface
	pool         *client.ShmPool
	buffer       *client.Buffer
	data         []byte
	tmpFile      *os.File
	done         bool
	runtimeErr   error
}

func main() {
	switch arg := firstArg(); arg {
	case "popup":
		must(togglePopup())
	case "serve-popup":
		must(servePopup())
	default:
		must(printModuleJSON())
	}
}

func firstArg() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	return ""
}

func printModuleJSON() error {
	now := time.Now()
	out := moduleOutput{Text: now.Format(moduleTimeFmt)}
	return json.NewEncoder(os.Stdout).Encode(out)
}

func togglePopup() error {
	if pid, ok := readPID(); ok {
		if err := syscall.Kill(pid, syscall.SIGTERM); err == nil {
			_ = os.Remove(pidFile)
			return nil
		}
		_ = os.Remove(pidFile)
	}

	self, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(self, "serve-popup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd.Start()
}

func servePopup() error {
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		return err
	}
	defer os.Remove(pidFile)

	app, err := newPopupApp()
	if err != nil {
		return err
	}
	defer app.close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		app.done = true
		_ = app.display.Context().Close()
	}()

	return app.run(time.Now())
}

func newPopupApp() (*popupApp, error) {
	display, err := client.Connect("")
	if err != nil {
		return nil, err
	}
	app := &popupApp{display: display}
	app.outputScale = 1
	app.preferred120 = 120
	display.SetErrorHandler(func(e client.DisplayErrorEvent) {
		app.fail(fmt.Errorf("wayland protocol error: code=%d message=%q", e.Code, e.Message))
	})

	registry, err := display.GetRegistry()
	if err != nil {
		_ = display.Context().Close()
		return nil, err
	}
	app.registry = registry
	registry.SetGlobalHandler(app.handleGlobal)

	if err := roundtrip(display); err != nil {
		app.close()
		return nil, err
	}
	if app.compositor == nil || app.shm == nil || app.layerShell == nil {
		app.close()
		return nil, errors.New("required Wayland globals are unavailable")
	}

	return app, nil
}

func (app *popupApp) run(now time.Time) error {
	if err := app.createSurface(now); err != nil {
		return err
	}

	for !app.done {
		if err := app.display.Context().Dispatch(); err != nil {
			if app.done {
				return app.runtimeErr
			}
			return err
		}
	}
	return app.runtimeErr
}

func (app *popupApp) handleGlobal(e client.RegistryGlobalEvent) {
	switch e.Interface {
	case "wl_compositor":
		if app.compositor == nil {
			obj := client.NewCompositor(app.display.Context())
			if err := app.registry.Bind(e.Name, e.Interface, minUint32(e.Version, 4), obj); err == nil {
				app.compositor = obj
			}
		}
	case "wl_shm":
		if app.shm == nil {
			obj := client.NewShm(app.display.Context())
			if err := app.registry.Bind(e.Name, e.Interface, 1, obj); err == nil {
				app.shm = obj
			}
		}
	case "wl_seat":
		if app.seat == nil {
			obj := client.NewSeat(app.display.Context())
			if err := app.registry.Bind(e.Name, e.Interface, minUint32(e.Version, 7), obj); err == nil {
				obj.SetCapabilitiesHandler(func(e client.SeatCapabilitiesEvent) {
					hasPointer := e.Capabilities&uint32(client.SeatCapabilityPointer) != 0
					if hasPointer && app.pointer == nil {
						pointer, err := obj.GetPointer()
						if err == nil {
							app.pointer = pointer
						}
					}
				})
				app.seat = obj
			}
		}
	case "wl_output":
		if app.output == nil {
			obj := client.NewOutput(app.display.Context())
			if err := app.registry.Bind(e.Name, e.Interface, minUint32(e.Version, 4), obj); err == nil {
				obj.SetScaleHandler(func(e client.OutputScaleEvent) {
					if e.Factor > 0 {
						app.outputScale = e.Factor
						if app.preferred120 == 120 {
							app.preferred120 = uint32(e.Factor) * 120
						}
					}
				})
				app.output = obj
			}
		}
	case "wp_viewporter":
		if app.viewporter == nil {
			obj := _jsii_viewporter.NewViewporter(app.display.Context())
			if err := app.registry.Bind(e.Name, e.Interface, 1, obj); err == nil {
				app.viewporter = obj
			}
		}
	case "wp_fractional_scale_manager_v1":
		if app.fractional == nil {
			obj := fractional_scale.NewFractionalScaleManager(app.display.Context())
			if err := app.registry.Bind(e.Name, e.Interface, 1, obj); err == nil {
				app.fractional = obj
			}
		}
	case "zwlr_layer_shell_v1":
		if app.layerShell == nil {
			obj := wlrlayershell.NewLayerShell(app.display.Context())
			if err := app.registry.Bind(e.Name, e.Interface, minUint32(e.Version, layerSurfaceVersion), obj); err == nil {
				app.layerShell = obj
			}
		}
	}
}

func (app *popupApp) createSurface(now time.Time) error {
	var err error
	app.surface, err = app.compositor.CreateSurface()
	if err != nil {
		return err
	}
	if app.viewporter != nil {
		app.viewport, err = app.viewporter.GetViewport(app.surface)
		if err != nil {
			return err
		}
	}
	if app.fractional != nil {
		app.surfaceScale, err = app.fractional.GetFractionalScale(app.surface)
		if err != nil {
			return err
		}
		app.surfaceScale.SetPreferredScaleHandler(func(e fractional_scale.FractionalScalePreferredScaleEvent) {
			if e.Scale > 0 {
				app.preferred120 = e.Scale
			}
		})
	}
	app.layerSurface, err = app.layerShell.GetLayerSurface(
		app.surface,
		app.output,
		uint32(wlrlayershell.LayerShellLayerTop),
		layerNamespace,
	)
	if err != nil {
		return err
	}

	app.layerSurface.SetConfigureHandler(func(e wlrlayershell.LayerSurfaceConfigureEvent) {
		if app.done {
			return
		}
		layout := newCalendarLayout(app.effectiveScale())
		img := app.renderCalendar(now, layout)
		if err := app.layerSurface.AckConfigure(e.Serial); err != nil {
			app.fail(err)
			return
		}
		if app.viewport != nil {
			if err := app.viewport.SetDestination(int32(layout.width), int32(layout.height)); err != nil {
				app.fail(err)
				return
			}
		}
		if app.buffer == nil {
			if err := app.createBuffer(layout, img); err != nil {
				app.fail(err)
				return
			}
		} else if err := app.replaceBuffer(layout, img); err != nil {
			app.fail(err)
			return
		}
		if err := app.surface.Attach(app.buffer, 0, 0); err != nil {
			app.fail(err)
			return
		}
		err := app.surface.DamageBuffer(0, 0, int32(layout.bufferWidth), int32(layout.bufferHeight))
		if err != nil {
			app.fail(err)
			return
		}
		if err := app.surface.Commit(); err != nil {
			app.fail(err)
		}
	})
	app.layerSurface.SetClosedHandler(func(wlrlayershell.LayerSurfaceClosedEvent) {
		app.fail(errors.New("compositor closed the calendar popup surface"))
	})

	anchor := uint32(wlrlayershell.LayerSurfaceAnchorBottom | wlrlayershell.LayerSurfaceAnchorRight)
	if err := app.layerSurface.SetAnchor(anchor); err != nil {
		return err
	}
	initialLayout := newCalendarLayout(app.effectiveScale())
	if err := app.layerSurface.SetSize(uint32(initialLayout.width), uint32(initialLayout.height)); err != nil {
		return err
	}
	if err := app.layerSurface.SetExclusiveZone(popupExclusiveZone); err != nil {
		return err
	}
	if err := app.layerSurface.SetMargin(0, popupRightMargin, popupBottomMargin, 0); err != nil {
		return err
	}
	if err := app.layerSurface.SetKeyboardInteractivity(uint32(wlrlayershell.LayerSurfaceKeyboardInteractivityNone)); err != nil {
		return err
	}
	if err := app.surface.SetBufferScale(1); err != nil {
		return err
	}

	return app.surface.Commit()
}

func (app *popupApp) createBuffer(layout calendarLayout, img *image.RGBA) error {
	if err := app.destroyBuffer(); err != nil {
		return err
	}
	size := layout.bufferWidth * layout.bufferHeight * 4
	tmpFile, err := os.CreateTemp("", "waybar-calendar-popup-*")
	if err != nil {
		return err
	}
	app.tmpFile = tmpFile
	if err := tmpFile.Truncate(int64(size)); err != nil {
		return err
	}

	data, err := syscall.Mmap(int(tmpFile.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	app.data = data

	pool, err := app.shm.CreatePool(int(tmpFile.Fd()), int32(size))
	if err != nil {
		return err
	}
	app.pool = pool

	buffer, err := pool.CreateBuffer(0, int32(layout.bufferWidth), int32(layout.bufferHeight), int32(layout.bufferWidth*4), uint32(client.ShmFormatXrgb8888))
	if err != nil {
		return err
	}
	app.buffer = buffer

	writeXRGB8888(data, img)
	return nil
}

func (app *popupApp) replaceBuffer(layout calendarLayout, img *image.RGBA) error {
	return app.createBuffer(layout, img)
}

func (app *popupApp) renderCalendar(now time.Time, layout calendarLayout) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, layout.bufferWidth, layout.bufferHeight))
	stddraw.Draw(img, img.Bounds(), &image.Uniform{C: colorBackground}, image.Point{}, stddraw.Src)

	borderRect := image.Rect(0, 0, layout.bufferWidth, layout.bufferHeight)
	drawBorder(img, borderRect, colorBorder, layout.borderWidth)

	months := []time.Time{
		time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location()),
		time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()),
	}
	monthXs := []int{layout.month1X, layout.month2X, layout.month3X}

	for idx, month := range months {
		x := monthXs[idx]
		drawCenteredString(img, layout.face, month.Format("January 2006"), x, layout.titleY, layout.monthWidth, colorText)
		drawString(img, layout.face, strings.Join(weekdayNames, " "), x, layout.weekdayY, colorMuted)

		cells := monthCells(month)
		for week := 0; week < len(cells); week++ {
			for weekday := 0; weekday < len(cells[week]); weekday++ {
				day := cells[week][weekday]
				if day == 0 {
					continue
				}
				cellX := x + weekday*layout.dayCellWidth
				cellY := layout.weeksTopY + week*layout.dayCellHeight
				text := fmt.Sprintf("%2d", day)
				if sameDay(now, month.Year(), month.Month(), day) {
					highlightRect := image.Rect(
						cellX-layout.highlightInsetX,
						cellY-layout.ascent+layout.highlightInsetY,
						cellX+font.MeasureString(layout.face, text).Ceil()+layout.highlightInsetX,
						cellY+layout.descent-layout.highlightInsetY,
					)
					stddraw.Draw(img, highlightRect, &image.Uniform{C: colorHighlight}, image.Point{}, stddraw.Src)
					drawString(img, layout.face, text, cellX, cellY, colorHighlightF)
					continue
				}
				drawString(img, layout.face, text, cellX, cellY, colorText)
			}
		}
	}

	return img
}

func newCalendarLayout(scale float64) calendarLayout {
	if scale < 1 {
		scale = 1
	}
	baseFace := loadPopupFace(1)
	baseCharWidth := font.MeasureString(baseFace, "M").Ceil()
	baseMetrics := baseFace.Metrics()
	baseLineHeight := baseMetrics.Height.Ceil() + 1
	baseMonthWidth := baseCharWidth * monthTitleWidth
	baseMonthGap := baseCharWidth * monthGapChars
	baseContentWidth := baseMonthWidth*3 + baseMonthGap*2
	baseContentHeight := baseLineHeight * 8
	width := baseContentWidth + popupPaddingLeft + popupPaddingRight
	height := baseContentHeight + popupPaddingY*2

	face := loadPopupFace(scale)
	charWidth := font.MeasureString(face, "M").Ceil()
	metrics := face.Metrics()
	lineHeight := metrics.Height.Ceil() + int(scale+0.5)
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	monthWidth := charWidth * monthTitleWidth
	monthGap := charWidth * monthGapChars
	paddingLeft := int(float64(popupPaddingLeft)*scale + 0.5)
	paddingY := int(float64(popupPaddingY)*scale + 0.5)
	titleY := paddingY + ascent
	weekdayY := paddingY + lineHeight + ascent
	weeksTopY := paddingY + lineHeight*2 + ascent
	dayCellWidth := charWidth * 3
	bufferWidth := int(float64(width)*scale + 0.5)
	bufferHeight := int(float64(height)*scale + 0.5)

	return calendarLayout{
		face:            face,
		scale:           scale,
		bufferWidth:     bufferWidth,
		bufferHeight:    bufferHeight,
		charWidth:       charWidth,
		lineHeight:      lineHeight,
		ascent:          ascent,
		descent:         descent,
		width:           width,
		height:          height,
		contentWidth:    baseContentWidth,
		contentHeight:   baseContentHeight,
		monthWidth:      monthWidth,
		monthGap:        monthGap,
		padding:         paddingY,
		borderWidth:     maxInt(1, int(float64(popupBorderWidth)*scale+0.5)),
		titleY:          titleY,
		weekdayY:        weekdayY,
		weeksTopY:       weeksTopY,
		month1X:         paddingLeft,
		month2X:         paddingLeft + monthWidth + monthGap,
		month3X:         paddingLeft + (monthWidth+monthGap)*2,
		dayCellWidth:    dayCellWidth,
		dayCellHeight:   lineHeight,
		highlightInsetX: maxInt(1, int(scale+0.5)),
		highlightInsetY: maxInt(1, int(scale+0.5)),
	}
}

func loadPopupFace(scale float64) font.Face {
	fontData, err := os.ReadFile(popupFontPath)
	if err != nil {
		return basicfont.Face7x13
	}
	parsed, err := opentype.Parse(fontData)
	if err != nil {
		return basicfont.Face7x13
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    popupFontSize * scale,
		DPI:     popupFontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return basicfont.Face7x13
	}
	return face
}

func monthCells(month time.Time) [6][7]int {
	var grid [6][7]int
	firstWeekday := int(month.Weekday())
	daysInMonth := daysIn(month)
	dayNum := 1
	for slot := firstWeekday; slot < firstWeekday+daysInMonth; slot++ {
		week := slot / 7
		weekday := slot % 7
		grid[week][weekday] = dayNum
		dayNum++
	}
	return grid
}

func daysIn(month time.Time) int {
	return time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, month.Location()).Day()
}

func sameDay(now time.Time, year int, month time.Month, day int) bool {
	return now.Year() == year && now.Month() == month && now.Day() == day
}

func drawCenteredString(img *image.RGBA, face font.Face, s string, x, baseline, width int, c color.Color) {
	textWidth := font.MeasureString(face, s).Ceil()
	drawString(img, face, s, x+(width-textWidth)/2, baseline, c)
}

func drawString(img *image.RGBA, face font.Face, s string, x, baseline int, c color.Color) {
	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, baseline),
	}
	d.DrawString(s)
}

func drawBorder(img *image.RGBA, rect image.Rectangle, c color.Color, width int) {
	top := image.Rect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Min.Y+width)
	bottom := image.Rect(rect.Min.X, rect.Max.Y-width, rect.Max.X, rect.Max.Y)
	left := image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y)
	right := image.Rect(rect.Max.X-width, rect.Min.Y, rect.Max.X, rect.Max.Y)
	stddraw.Draw(img, top, &image.Uniform{C: c}, image.Point{}, stddraw.Src)
	stddraw.Draw(img, bottom, &image.Uniform{C: c}, image.Point{}, stddraw.Src)
	stddraw.Draw(img, left, &image.Uniform{C: c}, image.Point{}, stddraw.Src)
	stddraw.Draw(img, right, &image.Uniform{C: c}, image.Point{}, stddraw.Src)
}

func writeXRGB8888(dst []byte, img *image.RGBA) {
	i := 0
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixel := uint32(r>>8)<<16 | uint32(g>>8)<<8 | uint32(b>>8)
			binary.LittleEndian.PutUint32(dst[i:i+4], pixel)
			i += 4
		}
	}
}

func roundtrip(display *client.Display) error {
	cb, err := display.Sync()
	if err != nil {
		return err
	}
	done := false
	cb.SetDoneHandler(func(client.CallbackDoneEvent) {
		done = true
		_ = cb.Destroy()
	})
	for !done {
		if err := display.Context().Dispatch(); err != nil {
			return err
		}
	}
	return nil
}

func (app *popupApp) close() {
	_ = app.destroyBuffer()
	if app.pointer != nil {
		_ = app.pointer.Release()
	}
	if app.seat != nil {
		_ = app.seat.Release()
	}
	if app.layerSurface != nil {
		_ = app.layerSurface.Destroy()
	}
	if app.surfaceScale != nil {
		_ = app.surfaceScale.Destroy()
	}
	if app.viewport != nil {
		_ = app.viewport.Destroy()
	}
	if app.surface != nil {
		_ = app.surface.Destroy()
	}
	if app.fractional != nil {
		_ = app.fractional.Destroy()
	}
	if app.viewporter != nil {
		_ = app.viewporter.Destroy()
	}
	if app.layerShell != nil {
		_ = app.layerShell.Destroy()
	}
	if app.output != nil {
		_ = app.output.Release()
	}
	if app.registry != nil {
		_ = app.registry.Destroy()
	}
	if app.data != nil {
		_ = syscall.Munmap(app.data)
	}
	if app.tmpFile != nil {
		name := app.tmpFile.Name()
		_ = app.tmpFile.Close()
		_ = os.Remove(name)
	}
	if app.display != nil {
		_ = app.display.Context().Close()
	}
}

func (app *popupApp) destroyBuffer() error {
	if app.buffer != nil {
		if err := app.buffer.Destroy(); err != nil {
			return err
		}
		app.buffer = nil
	}
	if app.pool != nil {
		if err := app.pool.Destroy(); err != nil {
			return err
		}
		app.pool = nil
	}
	if app.data != nil {
		if err := syscall.Munmap(app.data); err != nil {
			return err
		}
		app.data = nil
	}
	if app.tmpFile != nil {
		name := app.tmpFile.Name()
		if err := app.tmpFile.Close(); err != nil {
			return err
		}
		_ = os.Remove(name)
		app.tmpFile = nil
	}
	return nil
}

func (app *popupApp) effectiveScale() float64 {
	if app.preferred120 > 0 {
		return float64(app.preferred120) / 120.0
	}
	if app.outputScale > 0 {
		return float64(app.outputScale)
	}
	return 1
}

func (app *popupApp) fail(err error) {
	app.runtimeErr = err
	app.done = true
	if app.display != nil {
		_ = app.display.Context().Close()
	}
}

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func readPID() (int, bool) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	return pid, true
}

func must(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func init() {
	_ = os.MkdirAll(filepath.Dir(pidFile), 0o755)
}
