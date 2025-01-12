log_level = "info" # can be "debug", "info", "warn", "error", "fatal", "panic"

endpoint "ip" "bad_ips" {
    modes = ["http"] # can be "http", "socket" or both
    http_listener = ":8079" # the address to listen on
    socket_listener = "tcp://:8078" # the address to listen on
    http_base_path = "/" # the base path for the API
    file = "nz.csv" # can be a file or a URL. will respect HTTP_PROXY, HTTPS_PROXY, and NO_PROXY environment variables
    file_format = "plain" # a plain file is a list of IPv4, IPv6, CIDR, or a mix of them in a plain text file 
    inverted = false # if inverted is true, the response will be 0 for a hit, and 1 for a miss
    auto_reload = 300 # the interval (in minutes) to reload the file. 0 to disable
}

# endpoint "ip" "tor_exit_nodes" {
#     file = "https://check.torproject.org/torbulkexitlist"
#     file_format = "plain"
#     inverted = false
#     auto_reload = "1h"
# }

# endpoint "composite" "bad_ips_and_tor" {
#     endpoints = ["bad_ips", "tor_exit_nodes"]
#     operator = "and" # can be "and" or "or"
# }