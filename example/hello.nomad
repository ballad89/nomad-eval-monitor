job "helloapp" {

	datacenters = ["dc1"]

	update {

		stagger = "10s"

		max_parallel = 1
	}

	group "hello" {

		count = 1

		restart {

			attempts = 2
			interval = "1m"

			delay = "10s"

			mode = "fail"
		}


		task "hello" {

			driver = "docker"


			config {
                          image = "gerlacdt/helloapp:v0.1.0"
                          port_map {
                            http = 8080
                          }
                        }
			service {
				name = "${TASKGROUP}-service"
				tags = ["global", "hello", "urlprefix-hello.internal/"]
				port = "http"
				check {
				  name = "alive"
				  type = "http"
				  interval = "10s"
				  timeout = "3s"
				  path = "/health"
				}
			}

			resources {
				cpu = 500 # 500 MHz
				memory = 128 # 128MB
				network {
					mbits = 1
					port "http" {
					}
				}
			}

			logs {
			    max_files = 10
			    max_file_size = 15
			}

			kill_timeout = "10s"
		}
	}
}
