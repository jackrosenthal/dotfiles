;;; alan-mode.el --- major mode for Adventure LANguage (alanif.se) -*- lexical-binding: t -*-

;; Author: Jack Rosenthal
;; Maintainer: Jack Rosenthal
;; Version: 0.1
;; Package-Requires: (dependencies)
;; Homepage: homepage
;; Keywords: keywords


;; This file is not part of GNU Emacs

;; This file is free software; you can redistribute it and/or modify
;; it under the terms of the GNU General Public License as published by
;; the Free Software Foundation; either version 3, or (at your option)
;; any later version.

;; This program is distributed in the hope that it will be useful,
;; but WITHOUT ANY WARRANTY; without even the implied warranty of
;; MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
;; GNU General Public License for more details.

;; For a full copy of the GNU General Public License
;; see <http://www.gnu.org/licenses/>.


;;; Commentary:

;; commentary

;;; Code:

(defvar alan-mode-map
  (let ((map (make-keymap)))
    map)
  "Keymap for ALAN major mode")

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.alan\\'" . alan-mode))

(defconst alan-keywords
  '("Add" "After" "An" "And" "Are"
    "Article" "At" "Attributes" "Before" "Between"
    "By" "Can" "Cancel" "Character" "Characters"
    "Check" "Container" "Contains" "Count" "Current"
    "Decrease" "Definite" "Depend" "Depending" "Describe"
    "Description" "Directly" "Do" "Does" "Each"
    "Else" "ElsIf" "Empty" "End" "Entered"
    "Event" "Every" "Exclude" "Exit" "Extract"
    "First" "For" "Form" "From" "Has"
    "Header" "Here" "If" "Import" "In"
    "Include" "Increase" "Indefinite" "Indirectly" "Initialize"
    "Into" "Is" "IsA" "It" "Last"
    "Limits" "List" "Locate" "Look" "Make"
    "Max" "Mentioned" "Message" "Meta" "Min" "Name"
    "Near" "Nearby" "Negative" "No"
    "Not" "Of" "Off" "On" "Only"
    "Opaque" "Option" "Options" "Or" "Play"
    "Prompt" "Pronoun" "Quit" "Random" "Restart"
    "Restore" "Save" "Say" "Schedule" "Score"
    "Script" "Set" "Show" "Start" "Step"
    "Stop" "Strip" "Style" "Sum" "Synonyms"
    "Syntax" "System" "Taking" "The" "Then"
    "This" "To" "Transcript" "Transitively" "Until"
    "Use" "Verb" "Visits" "Wait" "When"))

(defconst alan-builtins
  '("entity" "thing" "location" "literal" "object" "actor" "string" "integer"
    "hero"))

(defun --alan-list-to-insensitive-re (words)
  (format "\\<\\(%s\\)\\>"
          (mapconcat (lambda (word)
                       (mapconcat (lambda (chr)
                                    (format "[%c%c]"
                                            (downcase chr)
                                            (upcase chr)))
                                  word ""))
                     words "\\|")))

(defconst alan-font-lock-keywords
  `(("\"\\([^\"]\\|\\\"\\)*\"" . font-lock-string-face)
    ("'\\([^']\\|''\\)*'" . font-lock-variable-name-face)
    ("--[^\n]*" . font-lock-comment-face)
    ("[0-9]+" . font-lock-constant-face)
    (,(--alan-list-to-insensitive-re alan-keywords)
     . font-lock-keyword-face)
    (,(--alan-list-to-insensitive-re alan-builtins)
     . font-lock-builtin-face)
    ("[A-Za-z][A-Za-z0-9_]*"
     . font-lock-variable-name-face)))

(defvar alan-mode-syntax-table
  (let ((st (make-syntax-table)))
    st))

(define-derived-mode alan-mode prog-mode "ALAN"
  "Major mode for editing Adventure LANguage source files"
  :syntax-table alan-mode-syntax-table
  (setq-local font-lock-defaults '(alan-font-lock-keywords))
  (setq-local font-lock-multiline t)
  (setq-local comment-start "--")
  (setq-local comment-end ""))

(provide 'alan-mode)

;;; alan-mode.el ends here
