" Lightweight VIMRC for when EMACS is not available

set nu
colorscheme murphy

set nocompatible

syn on
set t_Co=256  " make use of 256 terminal colors
set expandtab
set shiftwidth=4
set softtabstop=4
set browsedir=buffer
set scrolloff=5
set dir=~/.local/vimswap//
set undodir=~/.local/vimundo//
set undofile
set hidden
set spelllang=en_us
set ffs=unix,dos,mac

autocmd BufNewFile,BufRead *.tex setlocal tw=79
autocmd BufNewFile,BufRead *.cls setlocal tw=79
autocmd BufNewFile,BufRead *.cls setf tex
autocmd BufNewFile,BufRead *.tex setlocal spell

autocmd BufNewFile,BufRead *.md setlocal spell tw=79
autocmd BufNewFile,BufRead *.rst setlocal spell tw=79
autocmd BufNewFile,BufRead *.rb setlocal shiftwidth=2 softtabstop=2
autocmd BufNewFile,BufRead *.html setlocal shiftwidth=2 softtabstop=2
autocmd BufNewFile,BufRead *.htm setlocal shiftwidth=2 softtabstop=2
autocmd BufNewFile,BufRead *.css setlocal shiftwidth=2 softtabstop=2
autocmd BufNewFile,BufRead *.js setlocal shiftwidth=2 softtabstop=2
autocmd BufNewFile,BufRead *.blade.php setf html
autocmd BufNewFile,BufRead *.blade.php setlocal shiftwidth=2 softtabstop=2

autocmd BufNewFile,BufRead *.s colorscheme elmindreda

autocmd BufRead /tmp/mutt-* setlocal tw=72
autocmd BufRead /tmp/mutt-* setlocal spell spelllang=en_us

filetype plugin indent on
