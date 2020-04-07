;; -*- lexical-binding: t -*-
(require 'seq)
(require 'ivy)
(require 'request)

(defgroup monorail nil
  "Monorail Options")

(defcustom monorail-host "bugs.chromium.org"
  "The host to contact for monorail queries."
  :type 'string)

(defcustom monorail-user nil
  "The user to use for monorail queries."
  :type 'string)

(defvar monorail--cached-results nil)

(defun monorail--query-as-html (url cb)
  (cl-labels ((success-fcn (&key data &allow-other-keys)
                           (with-demoted-errors "Error in monorail callback: %S"
                             (funcall cb data))))
    (request url
      :parser (lambda ()
                (libxml-parse-html-region (point-min) (point-max) nil t))
      :success #'success-fcn)))

(defun monorail--get-issue-num-from-href (text)
  (and (string-match "id=\\([0-9]+\\)" text)
       (match-string 1 text)))

(defun monorail--dejunk-title (text)
  (and (string-match "(\\([^)]+\\))" text)
       (match-string 1 text)))

(defun monorail--filter-html-to-updates-title (html)
  (let ((a-elem (caddr html))
        (title-text-with-junk (cadddr html)))
    (cond
     ((and (stringp title-text-with-junk)
           (consp a-elem)
           (eq (car a-elem) 'a)
           (string-equal (cdr (assq 'class (cadr a-elem))) "ot-issue-link"))
      (let ((issue-num (monorail--get-issue-num-from-href
                        (cdr (assq 'href (cadr a-elem)))))
            (issue-title (monorail--dejunk-title title-text-with-junk)))
        (list (format "%s: %s" issue-num issue-title))))
     (t nil))))

(defun monorail--filter-html-to-updates (html)
  (cond
   ((stringp html) nil)
   ((and (consp html)
         (eq (car html) 'span)
         (string-equal (cdr (assq 'class (cadr html))) "title"))
    (monorail--filter-html-to-updates-title html))
   ((and (consp html)
         (eq (car html) 'head))
    nil)
   (t (cl-loop for elem in (cddr html)
               nconcing (monorail--filter-html-to-updates elem)))))

(defun monorail--query-updates ()
  (monorail--query-as-html
   (format "https://%s/u/%s/updates" monorail-host monorail-user)
   (lambda (html)
     (setq monorail--cached-results
           (delete-dups
            (nconc
             (monorail--filter-html-to-updates html)
             monorail--cached-results))))))

(defun monorail--ivy-candidates (input)
  (if monorail--cached-results
      (seq-filter (lambda (result)
                    (string-match-p input result))
                  monorail--cached-results)
    '("" "Querying monorail... type to update.")))

(defun monorail-insert-recent-ivy ()
  (interactive)
  (let ((prefix (if (string-match-p "^BUG=" (or (thing-at-point 'line t) ""))
                    "chromium:"
                  "crbug.com/"))
        (insertion-buffer (current-buffer)))
    (monorail--query-updates)
    (ivy-read
     "Monorail Bug: "
     #'monorail--ivy-candidates
     :action (lambda (bug-desc)
               (string-match "^\\([0-9]+\\)" bug-desc)
               (with-current-buffer insertion-buffer
                 (let ((bug-num (match-string 1 bug-desc)))
                   (insert (format "%s%s" prefix bug-num)))))
     :dynamic-collection t)))

(provide 'monorail)
