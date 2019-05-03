(custom-set-variables
 ;; custom-set-variables was added by Custom.
 ;; If you edit it by hand, you could mess it up, so be careful.
 ;; Your init file should contain only one such instance.
 ;; If there is more than one, they won't work right.
 '(column-number-mode t)
 '(debug-on-error nil)
 '(hl-paren-background-colors (quote ("gainsboro")))
 '(hl-paren-colors (quote ("dark cyan")))
 '(paren-face-modes
   (quote
    (lisp-mode emacs-lisp-mode lisp-interaction-mode ielm-mode scheme-mode inferior-scheme-mode clojure-mode cider-repl-mode nrepl-mode arc-mode inferior-arc-mode racket-mode)))
 '(paren-face-regexp "[][()]")
 '(whitespace-style
   (quote
    (face trailing tabs lines-tail empty space-before-tab tab-mark))))
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
