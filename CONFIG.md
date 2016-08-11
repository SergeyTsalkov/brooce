# Configuration
The first time brooce runs, it will create a `~/.brooce` dir in your home directory with a default `~/.brooce/brooce.conf` config file. That default config file is shown here, and we will explain what it all means in the following sections.

```json
{
  "cluster_name": "brooce",
  "timeout": 3600,
  "web": {
    "addr": ":8080",
    "certfile": "",
    "keyfile": "",
    "username": "admin",
    "password": "eoioszzi",
    "no_auth": false,
    "disable": false
  },
  "file_output_log": {
    "enable": false
  },
  "redis_output_log": {
    "drop_done": false,
    "drop_failed": false,
    "expire_after": 604800
  },
  "job_results": {
    "drop_done": false,
    "drop_failed": false
  },
  "redis": {
    "host": "localhost:6379",
    "password": ""
  },
  "suicide": {
    "enable": false,
    "command": "",
    "time": 0
  },
  "queues": {
    "common": 1
  },
  "path": ""
}
```

### cluster_name
Leave this alone unless you want multiple sets of workers to share one redis server. Multiple brooce workers on separate machines can normally draw jobs from the same queue, but putting them in separate clusters will make them unaware of each other.

### timeout
How long jobs can run before they're killed. Individual jobs can override this. The global default is 1 hour (3600 seconds). The timeout can't be disabled -- you should set it to the most time you expect your jobs to take, so it will automatically kill any that get stuck.
 
### web.addr
Where the web server is hosted. Defaults on port 8080 on all IPs that it can bind to.
 
### web.certfile / web.keyfile
Specify your HTTPS certificate and private key files here. If you have multiple certificate files, concatenate them into one file. If these are left blank, the web server will run in HTTP mode.
 
### web.username / web.password
We generate random login credentials the first time you run brooce, but you can change them here.
 
### web.no_auth
To run the web server with no authentication, leave username/password (above) blank, and set this to true. This is not recommended if you're having the web server listen on an internet-connected IP.
 
### web.disable
Set to true to disable the web server.
 
### file_output_log.enable
By default, job stdout/stderr is only logged to redis for review through the web interface. If you turn this on, the `~/.brooce` folder will get a logfile for every worker.
 
### redis_output_log.drop_done / redis_output_log.drop_failed
By default, we keep the logs for every job and store those logs in redis for access through the web interface. To save space, you can have those logs purged for jobs that succeed, or jobs that fail, or both.
 
### redis_output_log.expire_after
By default, job logs stored in redis expire after a week. You can change that here, in seconds.
 
### job_results.drop_done / job_results.drop_failed
By default, we store the name of each completed job in redis for later review through the web interface. You can drop those records for succeeded jobs, or failed jobs, or both.

### redis.host / redis.password
The hostname and password to access your redis server. Defaults to localhost and no-password.

### suicide.enable / suicide.command / suicide.time
For example, if you enabled suicide and set command to `"sudo shutdown -h now"` and time to `600`, you could shutdown your server after there haven't been any jobs for some time. Useful for shutting down idle EC2 instances. Keep in mind that the brooce program will need to have proper permissions to execute the given command, without additional prompts for passwords.

### queues
Brooce is multithreaded, and can listen for commands on multiple queues. For example, you could do the following to run 5 threads on the common queue and 2 more threads on the rare queue.
```json
{
  "queues": {
    "common": 5,
    "rare": 2
  }
}
```

### path
Add a given string to the brooce worker's PATH for running commands. For example, if you specify `"/home/mydir/bin"`, then now you can run a job as `mytask` instead of `/home/mydir/bin/mytask`.
 
