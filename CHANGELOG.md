# CHANGELOG

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/) and [Keep a Changelog](http://keepachangelog.com/).



## Unreleased
---

### New

### Changes

### Fixes

### Breaks

## 0.3.0 - (2022-10-28)
---

### New
* add maxPrice setting (heating off/lower if price is higher than maxPrice)

### Fixes
* day ahead price handling

## 0.2.0 - (2022-10-28)
---

### New
* add fallback mode (schedule)


## 0.1.0 - (2022-10-26)
---

### New
* add docker build, fix tests, add timezone support
* Threshold can be overridden by setting activeHours
* add -dryrun cmdline option (does not control relay when this is set)

### Changes
* use interfaces

### Fixes
* use 0.00 price when pricing is not available


