(custom-set-variables
 ;; custom-set-variables was added by Custom.
 ;; If you edit it by hand, you could mess it up, so be careful.
 ;; Your init file should contain only one such instance.
 ;; If there is more than one, they won't work right.
 '(clang-format-executable "/usr/bin/clang-format")
 '(column-number-mode t)
 '(debug-on-error nil)
 '(helm-minibuffer-history-key "M-p")
 '(hl-paren-background-colors '("gainsboro"))
 '(hl-paren-colors '("dark cyan"))
 '(magit-bury-buffer-function 'magit-mode-quit-window)
 '(paren-face-modes
   '(lisp-mode emacs-lisp-mode lisp-interaction-mode ielm-mode scheme-mode inferior-scheme-mode clojure-mode cider-repl-mode nrepl-mode arc-mode inferior-arc-mode racket-mode))
 '(paren-face-regexp "[][()]")
 '(safe-local-variable-values
   '((eval if
           (file-exists-p
            (concat
             (file-name-directory
              (let
                  ((d
                    (dir-locals-find-file ".")))
                (if
                    (stringp d)
                    d
                  (car d))))
             "/run_tests"))
           (set
            (make-local-variable 'compile-command)
            (if buffer-file-name
                (shell-quote-argument
                 (file-name-sans-extension buffer-file-name)))))))
 '(warning-suppress-log-types '((comp) (comp)))
 '(warning-suppress-types '((comp)))
 '(whitespace-style
   '(face trailing tabs lines-tail empty space-before-tab tab-mark)))
(custom-set-faces
 ;; custom-set-faces was added by Custom.
 ;; If you edit it by hand, you could mess it up, so be careful.
 ;; Your init file should contain only one such instance.
 ;; If there is more than one, they won't work right.
 '(default ((t (:background nil))))
 '(lisp-extra-font-lock-backquote ((t (:inherit font-lock-warning-face :foreground "dark orange"))))
 '(lisp-extra-font-lock-quoted ((t (:inherit font-lock-constant-face :foreground "dark orange"))))
 '(lisp-extra-font-lock-quoted-function ((t (:inherit font-lock-function-name-face :foreground "DodgerBlue1"))))
 '(lisp-extra-font-lock-special-variable-name ((t (:inherit font-lock-warning-face :foreground "cyan2")))))
