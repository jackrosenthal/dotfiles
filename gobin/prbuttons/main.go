// prbuttons prints a 2×3 sheet of 3-3/8" buttons on letter paper.
package main

import (
	"fmt"
	"io"
	"log/slog"
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
	pdf, err := os.CreateTemp("", "prbuttons*.pdf")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	pdf.Close()
	defer os.Remove(pdf.Name())

	slog.Info("generating print sheet")
	if err := printSheet(cli.ButtonFile, pdf.Name()); err != nil {
		return fmt.Errorf("printSheet: %w", err)
	}

	slog.Info("printing", "printer", cli.Printer)
	cmd := exec.Command("lpr", "-P", cli.Printer, pdf.Name())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lpr: %w\n%s", err, out)
	}
	return nil
}

// printSheet compiles a 2×3 button print sheet PDF via xelatex.
func printSheet(imgPath, output string) error {
	dir, err := os.MkdirTemp("", "prbuttons_tex")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	buttonFile := "button" + strings.ToLower(filepath.Ext(imgPath))
	if err := copyFile(imgPath, filepath.Join(dir, buttonFile)); err != nil {
		return err
	}

	tex := fmt.Sprintf(`\documentclass[letterpaper]{article}
%% 3-3/8" buttons: 2 cols x 3 rows
%% left/right = (8.5 - 2*3.375 - 0.25) / 2 = 0.75in
%% top/bottom = (11 - 3*3.375 - 2*0.25) / 2 = 0.1875in
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
\newcommand{\printbutton}{%%
  \begin{tikzpicture}
    \useasboundingbox (-1.6875in,-1.6875in) rectangle (1.6875in,1.6875in);
    \begin{scope}
      \clip (0,0) circle (\the\buttonradius);
      \node[inner sep=0pt] at (0,0) {%%
        \includegraphics[width=\buttondiameter]{%s}%%
      };
    \end{scope}
    \draw[line width=0.5pt] (0,0) circle (\the\buttonradius);
  \end{tikzpicture}%%
}
\begin{document}
\noindent\printbutton\hspace{0.25in}\printbutton\\[0.25in]
\printbutton\hspace{0.25in}\printbutton\\[0.25in]
\printbutton\hspace{0.25in}\printbutton\par
\end{document}`, buttonFile)

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
