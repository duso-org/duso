#!/usr/bin/env ruby
# frozen_string_literal: true
# Regex - core Regexp (stdlib/language builtin), single realistic ops on a real doc
doc = File.read('../docs/learning-duso.md', encoding: 'UTF-8')

start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
has_datastore = !doc.match(/datastore\(/).nil?
contains_ms = (Process.clock_gettime(Process::CLOCK_MONOTONIC) - start) * 1000

start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
words = doc.scan(/\w+/)
find_ms = (Process.clock_gettime(Process::CLOCK_MONOTONIC) - start) * 1000

puts "contains() on #{doc.bytesize}-byte doc in #{contains_ms}ms (found=#{has_datastore})"
puts "find() on #{doc.bytesize}-byte doc in #{find_ms}ms (#{words.length} matches)"
