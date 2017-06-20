SHELL := /bin/bash

.PHONY: tests
tests:
	for suite in tests/* ; do \
		for f in $$suite/* ; do \
			if [ $${f: -3} == ".go" ] ; then \
				go build -o $${f%???} $$f ; \
			fi \
		done \
	done
