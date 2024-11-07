
# Wyag in go

## Overview

This is an implementation for write your on git in to this this link for the [article](https://wyag.thb.lt/#)
it works up to the rm command but the check-ignore does not work well and writing to the index breaks git format

neither this or the wyag implementation work on repos that are cloned from github possibly due to differant formating
of index or config

## installation steps

i will add a make file but for now just run
``` bash
  go build main.go
```
