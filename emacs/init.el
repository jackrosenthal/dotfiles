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
  (evil-set-initial-state 'git-commit-mode 'insert)
  (add-hook 'git-commit-mode
            (lambda ()
              (yas-activate-extra-mode 'git-commit-mode)
              (when (chroot-file-p default-directory)
                (save-excursion
                  (unless (re-search-forward "Signed-off-by: " nil t)
                    (apply #'git-commit-signoff (git-commit-self-ident))))))))

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

(define-derived-mode ebuild-mode shell-script-mode "Ebuild"
  "Simple extension on top of `shell-script-mode'."
  (sh-set-shell "bash")
  (setq tab-width 4)
  (setq indent-tabs-mode t))

(add-to-list 'auto-mode-alist '("\\.\\(ebuild\\|eclass\\)\\'" . ebuild-mode))

;; Experimental Eshell hacks
(defun new-eshell ()
  "Make a new eshell"
  (interactive)
  (eshell t))

;; regular expression of commands which will run in chroot if
;; default-directory is in chroot
(setq in-chroot-default-directory-commands
      "^\\(.*/\\)?\\(make|gcc|g\\+\\+|repo|python\\(?:[23]\\(?:\\.[0-9]\\)?\\)?\\)$")
(fmakunbound 'eshell/make)

(defun cros-sdk-eshell-wrapper (fcn command args)
  (cl-labels
      ((in-sdk (command)
               (funcall fcn (chroot-file-path "src/chromium/depot_tools/cros_sdk")
                        `("--no-ns-pid" "--working-dir" "." "--" ,command ,@args))))
    (cond
     ((chroot-file-p command)
      (in-sdk command))
     ((and (chroot-file-p default-directory)
           (string-match in-chroot-default-directory-commands command))
      (in-sdk (match-string 0 command)))
     (t (funcall fcn command args)))))
(advice-add #'eshell-external-command :around #'cros-sdk-eshell-wrapper)
