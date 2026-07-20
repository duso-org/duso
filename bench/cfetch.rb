#!/usr/bin/env ruby
# Concurrent fetch benchmark against delay_server.du.
# WORKERS env sets the worker count (default 100), each making 5 requests.
require 'net/http'

WORKERS = (ENV['WORKERS'] || '100').to_i

errors = 0
mutex = Mutex.new

start = Time.now
threads = (1..WORKERS).map do
  Thread.new do
    5.times do
      begin
        Net::HTTP.start('127.0.0.1', 8399, open_timeout: 60, read_timeout: 60) do |h|
          res = h.request(Net::HTTP::Get.new('/delay'))
          mutex.synchronize { errors += 1 } if res.code.to_i != 200
        end
      rescue
        mutex.synchronize { errors += 1 }
      end
    end
  end
end
threads.each(&:join)
total = ((Time.now - start) * 1000).round

puts "cfetch #{WORKERS} workers x 5 reqs: #{total}ms, errors: #{errors}"
