nomad-eval-monitor
==================

*Majority of this functionality is now covered by [nomad deployments](https://www.nomadproject.io/docs/commands/deployment/status.html)*


Tool to monitor the progress of a nomad evaluation.
This tool was developed to facilitate automating the orchestration of nomad jobs.

- `id`: the evaluation id to be monitored.
- `timeout`: How long to wait for allocations to be placed and services passing healthchecks. Default `10s` (10 seconds). Can be any valid `time.Duration` e.g `2m` (2 minutes), `2h` (2 hours). This value should be a ceiling for how long you estimate it should take your service to successfully start up.



```bash
01:27 $ nomad run -detach example/hello.nomad
Job registration successful
Evaluation ID: 2ec19db1-7ac2-c313-55ca-854eb0774b0a

// Service failed to start up within 20s
$ ./nomad-eval-monitor -id 2ec19db1-7ac2-c313-55ca-854eb0774b0a -timeout "20s"
evaluation 2ec19db1-7ac2-c313-55ca-854eb0774b0a has 1 allocations
Allocation timed up: 473ac5dd-7c34-5448-e71d-8021b8e1bcc9
+----------+----------+-------+---------+--------+----------+--------------------------------+
|   JOB    | ALLOC-ID | TASK  |  STATE  | FAILED | RESTARTS |             EVENTS             |
+----------+----------+-------+---------+--------+----------+--------------------------------+
| helloapp | 473ac5dd | hello | pending | false  |        0 | 03/04/18 01:27:09 SAST |       |
|          |          |       |         |        |          | Received | Task received by    |
|          |          |       |         |        |          | client                         |
+----------+----------+-------+---------+--------+----------+--------------------------------+
+---------+--------+
| SERVICE | STATUS |
+---------+--------+
Timeout! Job did not finish running within deadline of 20.000000 seconds

// Service successfully started up in 20s
./nomad-eval-monitor -id 2ec19db1-7ac2-c313-55ca-854eb0774b0a -timeout "20s"
evaluation 2ec19db1-7ac2-c313-55ca-854eb0774b0a has 1 allocations
Allocation running: 473ac5dd-7c34-5448-e71d-8021b8e1bcc9
+----------+----------+-------+---------+--------+----------+--------------------------------+
|   JOB    | ALLOC-ID | TASK  |  STATE  | FAILED | RESTARTS |             EVENTS             |
+----------+----------+-------+---------+--------+----------+--------------------------------+
| helloapp | 473ac5dd | hello | running | false  |        0 | 03/04/18 01:29:37 SAST |       |
|          |          |       |         |        |          | Started | Task started by      |
|          |          |       |         |        |          | client 03/04/18 01:27:09 SAST  |
|          |          |       |         |        |          | | Driver | Downloading image   |
|          |          |       |         |        |          | gerlacdt/helloapp:v0.1.0       |
|          |          |       |         |        |          | 03/04/18 01:27:09 SAST |       |
|          |          |       |         |        |          | Received | Task received by    |
|          |          |       |         |        |          | client                         |
+----------+----------+-------+---------+--------+----------+--------------------------------+
+---------------+---------+
|    SERVICE    | STATUS  |
+---------------+---------+
| hello-service | passing |
+---------------+---------+
Done! All allocations for evaluation 2ec19db1-7ac2-c313-55ca-854eb0774b0a are running
```
