;; -*- lexical-binding: t -*-
(require 'package)
(require 'cl-lib)

(setq package-archives
      '(("gnu" . "https://elpa.gnu.org/packages/")
        ("org" . "http://orgmode.org/elpa/")
        ("melpa-stable" . "http://stable.melpa.org/packages/")
        ("melpa" . "http://melpa.org/packages/"))
      package-archive-priorities
      '(("melpa" . 10)
        ("melpa-stable" . 5)
        ("org" . 3)
        ("gnu" . 0)))
(package-initialize)

(defmacro thunk (&rest body)
  `(lambda () ,@body))

(unless (package-installed-p 'use-package)
  (package-refresh-contents)
  (package-install 'use-package))
(require 'use-package)
(setq use-package-always-ensure t)

(add-hook 'prog-mode-hook
          (thunk (display-line-numbers-mode 1)))
(tool-bar-mode -1)
(scroll-bar-mode -1)
(column-number-mode 1)
(show-paren-mode 1)
(setq inhibit-startup-screen t)
(setq ring-bell-function 'ignore)
(setq-default indent-tabs-mode nil)
(defalias 'yes-or-no-p 'y-or-n-p)
(set-default-font "Iosevka-12")

(defun default-buffer-predicate (buffer)
  (not (cl-member (buffer-name buffer)
                  '("*Completions*" "*Messages*" "*scratch*" "*Help*"
                    "*Buffer List*")
                  :test #'string-equal)))

(add-hook 'after-make-frame-functions
          (lambda (frame)
            (set-frame-font "Iosevka-12" nil (list frame))
            (set-frame-parameter frame 'buffer-predicate
                                 #'default-buffer-predicate)))

(defmacro togglef (param)
  `(lambda ()
     (interactive)
     (setq ,param (not ,param))
     (message "%s: %s" (quote ,param) ,param)))

(setq c-default-style "k&r"
      c-basic-offset 4)

(setq custom-file "~/.emacs.d/custom.el")

(unless (file-exists-p custom-file)
  (write-region "" nil custom-file))

(load custom-file)

(setq backup-directory-alist '(("." . "~/.local/emacs/backups"))
      backup-by-copying t
      version-control t
      delete-old-versions t
      kept-new-versions 20
      kept-old-versions 5)

(cl-loop for (pat . dir) in backup-directory-alist
         do (make-directory dir t))

;; load on corp machines only
;; (require 'google nil t)

(use-package undo-tree
  :config (global-undo-tree-mode))

(use-package blackboard-theme
  :config (load-theme 'blackboard t))

(use-package evil
  :init (setq evil-want-integration t
              evil-want-keybinding nil)
  :config (progn
            (evil-mode 1)
            (evil-define-key nil evil-insert-state-map
              (kbd "C-t") 'complete-symbol)
            (evil-define-key nil evil-normal-state-map
              (kbd "C-d") (togglef indent-tabs-mode))))

(use-package evil-collection
  :after evil
  :custom
  (evil-collection-company-use-tng nil)
  (evil-collection-setup-minibuffer t)
  :init (evil-collection-init))

(use-package evil-surround
  :config (global-evil-surround-mode 1))

(use-package racket-mode
  :after evil
  :config (evil-define-key 'normal racket-mode-map
            "gz" 'racket-run-and-switch-to-repl))

(use-package scribble-mode)

(use-package lispy
  :config (progn
            (lispy-set-key-theme '())
            (dolist (hook '(emacs-lisp-mode-hook
                            racket-mode-hook
                            racket-repl-mode-hook))
              (add-hook hook (thunk (lispy-mode 1))))
            (evil-define-key 'insert lispy-mode-map
              (kbd "(")   #'lispy-parens
              (kbd "[")   #'lispy-brackets
              (kbd "\"")  #'lispy-quotes
              (kbd ";")   #'lispy-comment
              (kbd ")")   #'lispy-right-nostring
              (kbd "]")   #'lispy-right-nostring
              (kbd "DEL") #'lispy-delete-backward)
            (setq lispy-left "[([]"
                  lispy-right "[])]")))

(use-package lispyville
  :config (progn
            (lispyville-set-key-theme
             '(operators
               c-w
               prettify
               text-objects
               atom-movement
               slurp/barf-cp
               commentary))
            (add-hook 'lispy-mode-hook #'lispyville-mode)
            (evil-define-key 'insert lispyville-mode-map
              (kbd "C-t") 'complete-symbol)
            (evil-define-key 'normal lispyville-mode-map
              (kbd "(") #'lispyville-insert-at-beginning-of-list
              (kbd ")") #'lispyville-insert-at-end-of-list)))

(use-package lisp-extra-font-lock
  :config (lisp-extra-font-lock-global-mode 1))

(use-package paren-face
  :config (global-paren-face-mode 1))

(use-package tex
  :ensure auctex
  :config (progn
            (setq TeX-auto-save t)
            (setq TeX-parse-self t)))

(use-package aggressive-indent
  :config (global-aggressive-indent-mode 1))

(use-package php-mode)

(use-package sublime-themes)

(use-package pdf-tools
  :config (pdf-tools-install))

(use-package pollen-mode
  :load-path "~/Dropbox/fun/pollen-mode")
