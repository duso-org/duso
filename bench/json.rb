#!/usr/bin/env ruby
# frozen_string_literal: true
# JSON round-trip - json module (stdlib since Ruby 1.9)
require 'json'

n = 5000
objs = []
(1..n).each do |i|
  objs << { id: i, name: "user-#{i}", active: i % 3 == 0, score: i * 1.5, tags: %w[a b c] }
end

start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
text = objs.to_json
encoded = (Process.clock_gettime(Process::CLOCK_MONOTONIC) - start) * 1000

start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
back = JSON.parse(text)
decoded = (Process.clock_gettime(Process::CLOCK_MONOTONIC) - start) * 1000

puts "JSON encode #{n} objects in #{encoded}ms, decode in #{decoded}ms"
