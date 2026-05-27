// prbuttons prints a 2×3 sheet of 3-3/8" buttons on letter paper.
// It lightens the blue stripe by 25% in HSL lightness before printing.
package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
)

var cli struct {
	ButtonFile string `arg:"" name:"button-file" help:"Button image or PDF to print." type:"existingfile"`
	Printer    string `short:"P" default:"Brother_HL-3170CDW_series" help:"Destination printer."`
}

func main() {
	kong.Parse(&cli)
	if err := run(); err != nil {
		slog.Error("failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	slog.Info("converting to PNG", "input", cli.ButtonFile)
	pngPath, pngCleanup, err := toPNG(cli.ButtonFile)
	if err != nil {
		return fmt.Errorf("toPNG: %w", err)
	}
	defer pngCleanup()

	slog.Info("lightening blue stripe")
	litPath, err := lightenBlue(pngPath, 0.25)
	if err != nil {
		return fmt.Errorf("lightenBlue: %w", err)
	}
	defer os.Remove(litPath)

	pdf, err := os.CreateTemp("", "prbuttons*.pdf")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	pdf.Close()
	defer os.Remove(pdf.Name())

	slog.Info("generating print sheet")
	if err := printSheet(litPath, pdf.Name()); err != nil {
		return fmt.Errorf("printSheet: %w", err)
	}

	slog.Info("printing", "printer", cli.Printer)
	cmd := exec.Command("lpr", "-P", cli.Printer, pdf.Name())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lpr: %w\n%s", err, out)
	}
	return nil
}

// toPNG returns a path to a PNG version of src.
// For PDF inputs it shells out to pdftoppm at 300 DPI.
func toPNG(src string) (path string, cleanup func(), err error) {
	switch strings.ToLower(filepath.Ext(src)) {
	case ".png", ".jpg", ".jpeg":
		return src, func() {}, nil
	case ".pdf":
		dir, err := os.MkdirTemp("", "prbuttons")
		if err != nil {
			return "", nil, err
		}
		cleanup := func() { os.RemoveAll(dir) }
		prefix := filepath.Join(dir, "btn")
		cmd := exec.Command("pdftoppm", "-r", "300", "-png", "-singlefile", src, prefix)
		if out, err := cmd.CombinedOutput(); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("pdftoppm: %w\n%s", err, out)
		}
		return prefix + ".png", cleanup, nil
	default:
		return "", nil, fmt.Errorf("unsupported file type %q", filepath.Ext(src))
	}
}

// lightenBlue boosts HSL lightness of blue-range pixels by boost (0–1).
// Targets hue 170°–230° with saturation > 0.5 and lightness > 0.25,
// which isolates the blue stripe without affecting dark navy text.
func lightenBlue(src string, boost float64) (string, error) {
	f, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	out := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			rf := float64(r) / 65535
			gf := float64(g) / 65535
			bf := float64(b) / 65535

			h, s, l := rgbToHSL(rf, gf, bf)
			if h >= 170 && h <= 230 && s > 0.5 && l > 0.25 {
				l = math.Min(1, l+boost)
			}
			rf, gf, bf = hslToRGB(h, s, l)

			out.SetNRGBA(x, y, color.NRGBA{
				R: uint8(math.Round(rf * 255)),
				G: uint8(math.Round(gf * 255)),
				B: uint8(math.Round(bf * 255)),
				A: uint8(a >> 8),
			})
		}
	}

	tmp, err := os.CreateTemp("", "prbuttons*.png")
	if err != nil {
		return "", err
	}
	if err := png.Encode(tmp, out); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil
}

func rgbToHSL(r, g, b float64) (h, s, l float64) {
	cmax := math.Max(math.Max(r, g), b)
	cmin := math.Min(math.Min(r, g), b)
	delta := cmax - cmin
	l = (cmax + cmin) / 2
	if delta == 0 {
		return 0, 0, l
	}
	if l < 0.5 {
		s = delta / (cmax + cmin)
	} else {
		s = delta / (2 - cmax - cmin)
	}
	switch cmax {
	case r:
		h = 60 * math.Mod((g-b)/delta, 6)
	case g:
		h = 60 * ((b-r)/delta + 2)
	case b:
		h = 60 * ((r-g)/delta + 4)
	}
	if h < 0 {
		h += 360
	}
	return
}

func hslToRGB(h, s, l float64) (r, g, b float64) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	r += m
	g += m
	b += m
	return
}

// printSheet compiles a 2×3 button print sheet PDF via xelatex.
func printSheet(imgPath, output string) error {
	dir, err := os.MkdirTemp("", "prbuttons_tex")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err := copyFile(imgPath, filepath.Join(dir, "button.png")); err != nil {
		return err
	}

	const tex = `\documentclass[letterpaper]{article}
% 3-3/8" buttons: 2 cols x 3 rows
% left/right = (8.5 - 2*3.375 - 0.25) / 2 = 0.75in
% top/bottom = (11 - 3*3.375 - 2*0.25) / 2 = 0.1875in
\usepackage[left=0.75in, right=0.75in, top=0.1875in, bottom=0.1875in]{geometry}
\usepackage{graphicx}
\usepackage{tikz}
\pagestyle{empty}
\setlength{\parskip}{0pt}
\setlength{\parindent}{0pt}
\setlength{\lineskip}{0pt}
\newlength{\buttonradius}
\setlength{\buttonradius}{1.6875in}
\newlength{\buttondiameter}
\setlength{\buttondiameter}{3.375in}
\newcommand{\printbutton}{%
  \begin{tikzpicture}
    \useasboundingbox (-1.6875in,-1.6875in) rectangle (1.6875in,1.6875in);
    \begin{scope}
      \clip (0,0) circle (\the\buttonradius);
      \node[inner sep=0pt] at (0,0) {%
        \includegraphics[width=\buttondiameter]{button.png}%
      };
    \end{scope}
    \draw[line width=0.5pt] (0,0) circle (\the\buttonradius);
  \end{tikzpicture}%
}
\begin{document}
\noindent\printbutton\hspace{0.25in}\printbutton\\[0.25in]
\printbutton\hspace{0.25in}\printbutton\\[0.25in]
\printbutton\hspace{0.25in}\printbutton\par
\end{document}`

	if err := os.WriteFile(filepath.Join(dir, "print.tex"), []byte(tex), 0o644); err != nil {
		return err
	}

	cmd := exec.Command("xelatex", "-interaction=nonstopmode", "print.tex")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("xelatex: %w\n%s", err, out)
	}

	return copyFile(filepath.Join(dir, "print.pdf"), output)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
