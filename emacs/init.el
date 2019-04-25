;; -*- lexical-binding: t -*-
(require 'package)
(require 'cl-lib)

(setq package-archives
      '(("gnu" . "https://elpa.gnu.org/packages/")
        ("org" . "http://orgmode.org/elpa/")
        ("melpa-stable" . "http://stable.melpa.org/packages/")
        ("melpa" . "http://melpa.org/packages/"))
      package-archive-priorities
      '(("org" . 15)
        ("melpa" . 10)
        ("melpa-stable" . 5)
        ("gnu" . 0)))
(package-initialize)

;; don't save selected packages as a custom variable
(defun package--save-selected-packages (&rest args)
  nil)

(defmacro thunk (&rest body)
  `(lambda () ,@body))

(unless (package-installed-p 'use-package)
  (package-refresh-contents)
  (package-install 'use-package))
(require 'use-package)
(setq use-package-always-ensure t)

;; Use the google package, if available
(require 'google nil t)

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
  (let ((name (buffer-name buffer)))
    (not
     (or (string-match-p "^magit-" name)
         (cl-member name
                    '("*Completions*" "*Messages*" "*scratch*" "*Help*"
                      "*Buffer List*")
                    :test #'string-equal)))))

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

(add-hook 'lisp-mode-hook
          (thunk (set (make-local-variable 'lisp-indent-function)
                      'common-lisp-indent-function)))

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

(use-package ivy
  :config (progn
            (ivy-mode 1)
            (setq ivy-use-virtual-buffers t
                  ivy-count-format "(%d/%d) ")))

(use-package counsel
  :config (counsel-mode 1))

(use-package racket-mode
  :after evil
  :config (progn
            (evil-define-key 'normal racket-mode-map
              "gz" #'racket-run-and-switch-to-repl
              "gs" #'racket-cycle-paren-shapes)
            (evil-define-key 'normal racket-repl-mode-map
              "gs" #'racket-cycle-paren-shapes)))

(use-package scribble-mode)

(use-package lispy
  :config (progn
            (lispy-set-key-theme '())
            (dolist (hook '(emacs-lisp-mode-hook
                            racket-mode-hook
                            racket-repl-mode-hook))
              (add-hook hook (thunk (lispy-mode 1))))
            (evil-define-key 'insert lispy-mode-map
              (kbd "(") #'lispy-parens
              (kbd "[") #'lispy-brackets
              (kbd "\"") #'lispy-quotes
              (kbd ";") #'lispy-comment
              (kbd ")") #'lispy-right-nostring
              (kbd "]") #'lispy-right-nostring
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

(use-package pollen-mode)

(use-package kotlin-mode
  :config (setq kotlin-tab-width 4))

(use-package yasnippet
  :config (yas-global-mode 1))

(use-package yasnippet-snippets
  :after yasnippet
  :config (yasnippet-snippets-initialize))

(use-package org)

(use-package magit
  :after evil yasnippet
  :config
  (global-set-key (kbd "C-,") #'magit-status)
  (evil-set-initial-state 'git-commit-mode 'insert)
  (add-hook 'git-commit-mode
            (thunk
             (yas-activate-extra-mode 'git-commit-mode)
             (when (string-prefix-p (expand-file-name "~/chromiumos")
                                    default-directory)
               (save-excursion
                 (unless (re-search-forward "Signed-off-by: " nil t)
                   (apply #'git-commit-signoff (git-commit-self-ident))))))))

(define-derived-mode ebuild-mode shell-script-mode "Ebuild"
  "Simple extension on top of shell-script-mode"
  (sh-set-shell "bash")
  (setq tab-width 4)
  (setq indent-tabs-mode t))

(add-to-list 'auto-mode-alist '("\\.\\(ebuild\\|eclass\\)\\'" . ebuild-mode))
