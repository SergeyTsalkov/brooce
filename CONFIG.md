# Configuration
The first time brooce runs, it will create a `~/.brooce` dir in your home directory with a default `~/.brooce/brooce.conf` config file. That default config file is shown here, and we will explain what it all means in the following sections.

```json
{
  "cluster_name": "brooce",
  "global_job_options": {
    "timeout": 3600,
    "maxtries": 1,
    "requeuedelayed": 60,
    "redislogexpireafter": 604800
  },
  "web": {
    "addr": ":8080",
    "certfile": "",
    "keyfile": "",
    "username": "admin",
    "password": "afwyczvk",
    "no_auth": false,
    "no_log": false,
    "disable": false
  },
  "file_output_log": {
    "enable": false
  },
  "redis": {
    "host": "localhost:6379",
    "socket": "",
    "password": "",
    "db": 0
  },
  "suicide": {
    "enable": false,
    "command": "",
    "time": 0
  },
  "queues": [
    {
      "name": "common",
      "workers": 1,
      "job_options": {}
    }
  ],
  "basepath": "",
  "path": ""
}
```

### `cluster_name`
Leave this alone unless you want multiple sets of workers to share one redis server. Multiple brooce workers on separate machines can normally draw jobs from the same queue, but putting them in separate clusters will make them unaware of each other.

### `global_job_options`
Job options specified here will apply globally, unless overwritten in the queue job_options hash, or individually in the job. [See the list of all job options in README.md](README.md#job-options).
 
### `web.addr`
Where the web server is hosted. Defaults on port 8080 on all IPs that it can bind to.
 
### `web.certfile` / `web.keyfile`
Specify your HTTPS certificate and private key files here. If you have multiple certificate files, concatenate them into one file. If these are left blank, the web server will run in HTTP mode.
 
### `web.username` / `web.password`
We generate random login credentials the first time you run brooce, but you can change them here.
 
### `web.no_auth`
To run the web server with no authentication, leave username/password (above) blank, and set this to true. This is not recommended if you're having the web server listen on an internet-connected IP.
 
### `web.no_log`
By default, web interface access is logged to `~/.brooce/web.log`. Set this to true to disable web access logging.
 
### `web.disable`
Set to true to disable the web server.
 
### `file_output_log.enable`
By default, job stdout/stderr is only logged to redis for review through the web interface. If you turn this on, the `~/.brooce` folder will get a logfile for every worker.

### `redis.host` / `redis.password`
The hostname and password to access your redis server. Defaults to localhost and no-password.

### `redis.socket`
If specified, connect through a unix sock file instead of a hostname. Example: `/var/run/redis/redis-server.sock`

### `redis.db`
The db which will be used by brooce on your redis server. Defaults to 0.

### `suicide.enable` / `suicide.command` / `suicide.time`
For example, if you enabled suicide and set command to `"sudo shutdown -h now"` and time to `600`, you could shutdown your server after there haven't been any jobs for some time. Useful for shutting down idle EC2 instances. Keep in mind that the brooce program will need to have proper permissions to execute the given command, without additional prompts for passwords.

### `queues`
Brooce is multithreaded, and can listen for commands on multiple queues. For example, you could do the following to run 5 threads on the common queue and 2 more threads on the rare queue.

You can set per-queue job options in the `job_options` hash, as shown below. Per-queue options override `global_job_options`, and individual jobs can override the per-queue settings. [See the list of all job options in README.md](README.md#job-options).

```json
{
  "queues": [
    {
      "name": "common",
      "workers": 5
    },
    {
      "name": "rare",
      "workers": 2,
      "job_options": {
        "timeout": 7200,
        "maxtries": 5,
        "killondelay": true
      }
    }
  ]
}
```

### `basepath`
If you're running brooce's web server behind a proxy and exposing it as a folder (e.g. yourdomain.com/brooce/), you can set this to the directory path (e.g. `/brooce`).

### `path`
Add a given string to the brooce worker's PATH for running commands. For example, if you specify `/home/mydir/bin`, then now you can run a job as `mytask` instead of `/home/mydir/bin/mytask`.
 
