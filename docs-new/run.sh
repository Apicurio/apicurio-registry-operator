#!/bin/bash

# Build search index
bundle exec just-the-docs rake search:init

# Serve
bundle exec jekyll serve
