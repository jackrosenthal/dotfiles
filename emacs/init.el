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
;; (keep the file clean!)
(defun package--save-selected-packages (&rest args)
  nil)

;; Bootstrap use-package
(unless (package-installed-p 'use-package)
  (package-refresh-contents)
  (package-install 'use-package))
(require 'use-package)
(setq use-package-always-ensure t)

;; Setup environment
(defun add-to-path (dir)
  (let ((path (expand-file-name dir)))
    (unless (member path exec-path)
      (push path exec-path))
    (unless (member path (split-string (getenv "PATH") ":"))
      (setenv "PATH" (format "%s:%s" path (getenv "PATH"))))))

(add-to-path "~/dotfiles/bin")
(add-to-path "~/.local/bin")
(add-to-path "~/depot_tools")
(add-to-path "~/chromiumos/chromite/bin")

(setenv "PAGER" "cat")

;; Use the Google package, if available
(require 'google nil t)
(require 'google-logo nil t)
(setq google-use-coding-style nil)

(defvar cros-chroot-trunk "~/chromiumos")

(defun chroot-file-path (relative-path)
  (expand-file-name
   (concat (file-name-as-directory (expand-file-name cros-chroot-trunk))
           relative-path)))

(defun chroot-file-p (path)
  (string-prefix-p (chroot-file-path ".")
                   (expand-file-name path)))

;; EC hook information
(defun ec-hooktypes (enum-name)
  (with-temp-buffer
    (insert-file-contents (chroot-file-path "src/platform/ec/include/hooks.h"))
    (search-forward (format "enum %s {" enum-name) nil nil)
    (delete-region (point-min) (point))
    (search-forward "};" nil nil)
    (delete-region (point) (point-max))
    (goto-char (point-min))
    (cl-loop while (re-search-forward "HOOK_[A-Z_]+" nil t)
             unless (cl-member (match-string 0) result :test #'string=)
             collect (match-string 0) into result
             finally (return result))))

;; Display line numbers
(add-hook 'prog-mode-hook #'display-line-numbers-mode)

;; Whitespace mode
(defun maybe-activate-whitespace-mode ()
  "Activates `whitespace-mode' only if the buffer is from a file."
  (when (buffer-file-name)
    (whitespace-mode 1)))

(cl-loop for hook in '(text-mode-hook prog-mode-hook)
         do (add-hook hook #'maybe-activate-whitespace-mode))

;; Default interface setup
(tool-bar-mode -1)
(scroll-bar-mode -1)
(column-number-mode 1)
(show-paren-mode 1)
(setq inhibit-startup-screen t)
(setq ring-bell-function 'ignore)
(setq-default indent-tabs-mode nil)
(defalias 'yes-or-no-p 'y-or-n-p)
(set-frame-font "Iosevka-12")

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

;; Default C style configuration
(setq c-default-style "linux"
      c-basic-offset 8)

(add-hook 'c-mode-hook (lambda ()
                         (setq tab-width 8)
                         (setq indent-tabs-mode t)))

(when (fboundp #'google-set-c-style)
  (add-hook 'c++-mode-hook #'google-set-c-style))

;; Default Python configuration
(setq-default python-shell-interpreter "python3")

;; Place customize variables in a different file
(setq custom-file "~/.emacs.d/custom.el")
(load custom-file)

;; Automatic backup
(setq backup-directory-alist '(("." . "~/.local/emacs/backups"))
      backup-by-copying t
      version-control t
      delete-old-versions t
      kept-new-versions 20
      kept-old-versions 5)

(cl-loop for (pat . dir) in backup-directory-alist
         do (make-directory dir t))

;; Use Common Lisp standard indentation (rather than Emacs-style
;; indentation) in when editing Common Lisp
(add-hook 'lisp-mode-hook
          (lambda ()
            (set (make-local-variable 'lisp-indent-function)
                 'common-lisp-indent-function)))
(setq inferior-lisp-program "sbcl")

;; My package choices
(use-package clang-format)

(use-package undo-tree
  :config (global-undo-tree-mode))

(use-package afternoon-theme
  :config (load-theme 'afternoon t))

(use-package evil
  :init (setq evil-want-integration t
              evil-want-keybinding nil)
  :config
  (evil-mode 1)
  (evil-define-key nil evil-insert-state-map
    (kbd "C-t") 'complete-symbol))

(use-package evil-collection
  :after evil
  :custom
  (evil-collection-company-use-tng nil)
  (evil-collection-setup-minibuffer t)
  :init (evil-collection-init))

(use-package evil-surround
  :config (global-evil-surround-mode 1))

(use-package ivy
  :config
  (ivy-mode 1)
  (setq ivy-use-virtual-buffers t
        ivy-count-format "(%d/%d) "))

(use-package cmake-mode)

(use-package counsel
  :config (counsel-mode 1))

(use-package racket-mode
  :after evil
  :config
  (evil-define-key 'normal racket-mode-map
    "gz" #'racket-run-and-switch-to-repl
    "gs" #'racket-cycle-paren-shapes)
  (evil-define-key 'normal racket-repl-mode-map
    "gs" #'racket-cycle-paren-shapes))

(use-package scribble-mode)

(use-package lispy
  :config
  (lispy-set-key-theme '())
  (dolist (hook '(emacs-lisp-mode-hook
                  racket-mode-hook
                  racket-repl-mode-hook
                  lisp-mode-hook
                  ielm-mode-hook))
    (add-hook hook #'lispy-mode))
  (evil-define-key 'insert lispy-mode-map
    (kbd "(") #'lispy-parens
    (kbd "[") #'lispy-brackets
    (kbd "\"") #'lispy-quotes
    (kbd ";") #'lispy-comment
    (kbd ")") #'lispy-right-nostring
    (kbd "]") #'lispy-right-nostring
    (kbd "DEL") #'lispy-delete-backward)
  (setq lispy-left "[([]"
        lispy-right "[])]"))

(use-package lispyville
  :config
  (lispyville-set-key-theme
   '(operators
     c-w
     prettify
     text-objects
     atom-movement
     slurp/barf-cp
     commentary))
  (add-hook 'lispy-mode-hook #'lispyville-mode)
  (evil-define-key 'normal lispyville-mode-map
    (kbd "(") #'lispyville-insert-at-beginning-of-list
    (kbd ")") #'lispyville-insert-at-end-of-list))

(use-package lisp-extra-font-lock
  :config (lisp-extra-font-lock-global-mode 1))

(use-package paren-face
  :config (global-paren-face-mode 1))

(use-package smartparens
  :config
  (add-hook 'c-mode-common-hook #'smartparens-mode))

(use-package tex
  :ensure auctex
  :config
  (setq TeX-auto-save t)
  (setq TeX-parse-self t))

(use-package aggressive-indent
  :config (global-aggressive-indent-mode 1))

(use-package php-mode)

(use-package pdf-tools
  :config (pdf-tools-install))

(use-package pollen-mode)

(use-package kotlin-mode
  :config (setq kotlin-tab-width 4))

(use-package lua-mode)

(use-package rust-mode)

(use-package yasnippet
  :config (yas-global-mode 1))

(use-package yasnippet-snippets
  :after yasnippet
  :config
  (yasnippet-snippets-initialize)
  (setq yas-wrap-around-region t))

(use-package org)

(use-package magit
  :after evil yasnippet
  :config
  (global-set-key (kbd "C-,") #'magit-status)
  (define-key magit-mode-map (kbd "R") #'repo-main-menu)
  (evil-set-initial-state 'git-commit-mode 'insert)
  (add-hook 'git-commit-mode
            (lambda ()
              (yas-activate-extra-mode 'git-commit-mode)
              (when (chroot-file-p default-directory)
                (save-excursion
                  (unless (re-search-forward "Signed-off-by: " nil t)
                    (apply #'git-commit-signoff (git-commit-self-ident))))))))

(use-package ag)

(use-package diff-hl
  :after magit
  :config
  (global-diff-hl-mode)
  (add-hook 'magit-post-refresh-hook #'diff-hl-magit-post-refresh))

(use-package with-editor
  :config
  (cl-loop for mode in '(shell-mode-hook term-exec-hook eshell-mode-hook)
           do (add-hook mode #'with-editor-export-editor)))

(use-package page-break-lines
  :config (global-page-break-lines-mode))

(use-package gn-mode)

(use-package yaml-mode)

(use-package kconfig-mode
  :config
  (add-hook 'kconfig-mode-hook
            (lambda ()
              (setq tab-width 8)
              (setq indent-tabs-mode t))))

(define-derived-mode ebuild-mode shell-script-mode "Ebuild"
  "Simple extension on top of `shell-script-mode'."
  (sh-set-shell "bash")
  (setq tab-width 4)
  (setq indent-tabs-mode t))

(add-to-list 'auto-mode-alist '("\\.\\(ebuild\\|eclass\\)\\'" . ebuild-mode))

;; Eshell Setup
(require 'eshell)
(require 'em-smart)

(setq eshell-where-to-jump 'begin)
(setq eshell-review-quick-commands nil)
(setq eshell-smart-space-goes-to-end t)

(defun git-split-args (args options-one)
  (cl-labels ((rec (args-to-process partial-options)
                   (cond
                    ((not args-to-process) (cl-values (reverse partial-options)
                                                      nil))
                    ((string-equal "--" (car args-to-process))
                     (cl-values (reverse partial-options)
                                (cdr args-to-process)))
                    ((member (car args-to-process) options-one)
                     (rec (cddr args-to-process)
                          (list* (cadr args-to-process)
                                 (car args-to-process)
                                 partial-options)))
                    ((string-prefix-p "-" (car args-to-process))
                     (rec (cdr args-to-process)
                          (cons (car args-to-process) partial-options)))
                    (t (cl-values (reverse partial-options) args-to-process)))))
    (rec args '())))

(defun new-eshell ()
  "Make a new eshell"
  (interactive)
  (eshell t))

(defun eshell/git (&rest args)
  (cl-labels ((defer-external ()
                (eshell/wait
                 (eshell-external-command "git" `("--no-pager" ,@args)))))
    (cl-multiple-value-bind (git-args subcommand-and-args)
        (git-split-args args '("-C"))
      (let* ((chdir-arg (member "-C" git-args))
             (subcommand (car subcommand-and-args))
             (subcommand-args (cdr subcommand-and-args))
             (subcommand-fcn-symbol (intern (format "eshell/git-%s" subcommand))))
        (cond ((or (member "--version" git-args)
                   (member "--help" git-args)
                   (member "--html-path" git-args)
                   (member "--man-path" git-args)
                   (member "--info-path" git-args))
               (defer-external))
              (chdir-arg
               (let ((default-directory (expand-file-name (cadr chdir-arg))))
                 (apply #'eshell/git
                        `(,@(seq-take-while (lambda (arg)
                                              (not (eq arg (car chdir-arg))))
                                            git-args)
                          ,@(cddr git-args)
                          ,@subcommand-and-args))))
              ((not subcommand) (magit-status))
              ((and subcommand (fboundp subcommand-fcn-symbol))
               (apply subcommand-fcn-symbol subcommand-args))
              (t (defer-external)))))))

(defun eshell/git-status (&rest args)
  (magit-status))

(defun eshell/git-log (&rest args)
  (cl-multiple-value-bind (options rest)
      (git-split-args args '("-L" "-n" "-l" "-O"))
    (magit-log-setup-buffer (if rest
                                (car rest)
                              '("HEAD"))
                            options
                            (cdr rest))))

(defun eshell/take (directory)
  (let ((dir (expand-file-name directory)))
    (make-directory dir t)
    (setq default-directory dir)))

(defun cros-sdk-eshell-hook (command args)
  (cl-labels
      ((in-sdk (command)
               (throw 'eshell-replace-command
                      (eshell-parse-command
                       "cros_sdk"
                       `("--no-ns-pid" "--working-dir" "." "--" ,command ,@args)))))
    (when (string-prefix-p "@" command)
      (in-sdk (substring command 1)))))
(add-hook 'eshell-named-command-hook #'cros-sdk-eshell-hook)

;; Local Packages
(push "~/.emacs.d/site-lisp" load-path)
(require 'upstart-mode)
(require 'monorail)
(setq monorail-user "jrosenth@chromium.org")

(evil-define-key nil evil-insert-state-map
  (kbd "C-b") #'monorail-insert-recent-ivy)

(push "~/chromiumos/platform/dev/contrib/emacs" load-path)
(require 'repo-transient nil t)
