#!/usr/bin/env ruby
# frozen_string_literal: true
# Ruby benchmark server: GET /delay responds after 1 second,
# GET /ping responds immediately. Stdlib only - WEBrick was removed from
# core stdlib in Ruby 3.0 and never replaced, so this uses the modern
# Socket.tcp_server_loop idiom with a thread per connection.
require 'socket'

RESPONSE = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 2\r\nConnection: keep-alive\r\n\r\nok"

Socket.tcp_server_loop('127.0.0.1', 8399) do |conn|
  conn.setsockopt(Socket::IPPROTO_TCP, Socket::TCP_NODELAY, 1)
  Thread.new(conn) do |c|
    begin
      loop do
        request_line = c.gets
        break if request_line.nil?

        path = request_line.split[1]
        while (header = c.gets)
          break if header.strip.empty?
        end

        sleep 1 if path&.start_with?('/delay')
        c.write(RESPONSE)
      end
    rescue StandardError
      # client disconnected mid-request; nothing to clean up
    ensure
      c.close
    end
  end
end
