# brooce
Hey Hackers! Brooce is a **language-agnostic job queue** I made in Go. I built it because I like to work on personal projects in a variety of languages, and I want to be able to **use the same job queue regardless of what language I'm writing in**. I like a lot about Resque, but it has the same flaw as many others: you're all-but-forced to write jobs in its preferred language, Ruby.

Therefore, I built a job queue system where **the jobs themselves are just shell commands**. It's really simple to get started: you just grab the brooce binary and run it on any Linux system. You then use redis to LPUSH some shell commands to a queue, and then brooce will run them in sequence.

That's really all you need to know to use it, but there are some advanced features under the hood. There's a resque-inspired web interface, multi-threaded job execution, locking, and automatically scheduled cron-like jobs. All features are baked into a single binary that runs on any Linux platform, and can be deployed on an unlimited number of servers. If they can all access the same redis database, they'll all coordinate amongst themselves to work on jobs.

I've been personally relying on brooce with great results! If you try it out, I would welcome your feedback!

## Features

* **Single Executable** -- Brooce comes as a single executable that runs on any Linux system.
* **Redis Backend** -- Redis can be accessed from any programming language, or the command line. Schedule jobs from anywhere.
* **Language-Agnostic** -- Jobs are just shell commands. Write jobs in any language.
* **Scalable** -- Deploy instances on one server, or many. Each instance can run multiple jobs simultaneously. All instances coordinate amongst themselves.
* **Crash Recovery** -- If you run multiple instances of brooce on different servers, they'll monitor each other. All features can survive instances failures, and any jobs being worked on by crashed instances will be marked as failed.
* **Web Interface** -- Brooce runs its own password-protected web server. You can access it to monitor currently running jobs across all instances, and list jobs that are pending, delayed, done, or failed. You can look at the stdout/stderr output of jobs while they're running, or after they're done.
* **Job Logging** -- Job stdout/stderr output can be logged to redis or log files, for later review through the web interface or your favorite text editor.
* **Timeouts** -- To prevent jobs from getting stuck, brooce automatically kills any jobs that don't finish in an hour. You can change this default timeout in [brooce.conf](CONFIG.md), or set per-job timeouts.
* **Locking** -- Jobs can use brooce's lock system, or implement their own. A job that can't grab a lock it needs will be delayed and put back on the queue a minute later.
* **Cron Jobs** -- Schedule tasks to run on a schedule.
* **Suicide Mode** -- Instruct brooce to run a shell command after it's been idle for a pre-set period. Perfect for having unneeded EC2 workers terminate themselves.

## Learn Redis First
Brooce uses redis as its database. Redis can be accessed from any programming language, but how to do it for each one is beyond the scope of this documentation. All of our examples will use the redis-cli shell commands, and it's up to you to substitute the equavalents in your language of choice! If you're a programmer and you haven't learned redis yet, you owe it to yourself to do so!

## Quick Start
We have a tutorial that'll get you set up and run your first job.

[View Quick Start Tutorial](QUICKSTART.md)

 
## Configuration
The first time brooce runs, it will create a `~/.brooce` dir in your home directory with a default `~/.brooce/brooce.conf` config file. 

[View brooce.conf Documentation](CONFIG.md)
 






## Scheduling Jobs
By default, there is a single queue called "common" that uses a single thread to run jobs. To schedule a job on that queue, run: 
```shell
redis-cli LPUSH brooce:queue:common:pending 'ls -l ~ | tee ~/files.txt'
```

## Setting Up Multiple Queues
Brooce is multi-threaded, and can run many jobs at once on multiple queues. To set up multiple queues, edit the [queues section of brooce.conf](CONFIG.md#queues).

## Timeouts
So far, we've treated jobs as strings, but they can also be json hashes with additional parameters. Here is a job that overwrites the [default 1-hour timeout in brooce.conf](CONFIG.md#timeout) and runs for only 10 seconds:
```shell
redis-cli LPUSH brooce:queue:common:pending '{"command":"sleep 11 && touch ~/done.txt","timeout":10}'
```
In this example, the done.txt file will never be created because the job will be killed too soon. If you go into the web interface, you'll be able to see it under failed jobs.


## Locking
Locks can prevent multiple concurrent jobs from breaking things by touching the same resource at the same time. Let's say you have several kinds of jobs that touch a single account, and you don't want them to interfere with each other by running at the same time. You might schedule:
```shell
redis-cli LPUSH brooce:queue:common:pending '{"command":"~/bin/reconfigure-account.sh 671","locks":["account:671"]}'
redis-cli LPUSH brooce:queue:common:pending '{"command":"~/bin/bill-account.sh 671","locks":["account:671"]}'
```
Even if there are multiple workers available, only one of these jobs will run at a time. The other will get pushed into the delayed queue, which you can see in the web interface. Once per minute, the contents of the delayed queue are dumped back into the pending queue, where it'll get the chance to run again if it can grab the needed lock.

### Multiple Locks
You can pass multiple locks. Your job must grab all the locks to run:
```shell
redis-cli LPUSH brooce:queue:common:pending '{"command":"~/bin/reconfigure-account.sh 671","locks":["account:671","server:5"]}'
```

### Locks That Multiple Jobs Can Hold
A lock that begins with a number followed by a colon can be held by that many jobs at once. For example, let's say each server can tolerate no more than 3 jobs acting on it at once. You might run:
```shell
redis-cli LPUSH brooce:queue:common:pending '{"command":"~/bin/reconfigure-account.sh 671","locks":["account:671","3:server:5"]}'
```
The `account:671` lock must be exclusively held by this job, but the `3:server:5` lock means that up to 3 jobs can act on server 5 at the same time.

### Locking Things Yourself
Sometimes you don't know which locks a job will need until after it starts running -- maybe you have a script called `~/bin/bill-all-accounts.sh` and you want it to lock all accounts that it's about to bill. In that case, your script will need to implement its own locking system. If it determines that it can't grab the locks it needs, it should return exit code 75 (temp failure). All other non-0 exit codes cause your job to be marked as failed, but 75 causes it to be pushed to the delayed queue and later re-tried.


## Cron Jobs
Cron jobs work much the same way they do on Linux, except you're setting them up as redis keys and specifying a queue to run in. Let's say you want to bill all your users every day at midnight. You might do this:
```shell
redis-cli SET "brooce:cron:jobs:daily-biller" "0 0 * * * queue:common ~/bin/bill-all-accounts.sh"
```
**Cron job times are always UTC, regardless of your local time zone!** This was unavoidable since brooce instances could be running on multiple servers in different time zones.

You can see any pending cron jobs on the Cron Jobs page in the web interface.


### Timeouts and Locking in a Cron Job
Timeouts and locking are available to cron jobs. Here is an example that uses both:
```shell
redis-cli SET "brooce:cron:jobs:daily-biller" "0 0 * * * queue:common timeout:600 locks:server:5,server:8 ~/bin/bill-all-accounts.sh"
```
In this case, we want `~/bin/bill-all-accounts.sh` to run daily, finish in under 10 minutes, and hold locks on `server:5` and `server:8`.

### Don't Schedule Cron Job If It's Already Running
Let's say you want to run a job once per minute, but not schedule another one if the last one is still running. This is different than a lock, which would cause additional jobs to be delayed and pile up. There is a `skipifrunning` option that will do this.
```shell
# Clear cache once per minute, but don't schedule if the last one is still running
redis-cli SET "brooce:cron:jobs:clear-cache" "* * * * * queue:common skipifrunning:true ~/bin/clear-cache.sh"
```


### Fancy Cron Jobs
Most of the standard cron features are implemented. Here are some examples.
```shell
# Bill accounts twice a day
redis-cli SET "brooce:cron:jobs:daily-biller" "0 */12 * * * queue:common ~/bin/bill-all-accounts.sh"

# Rotate logs 4 times an hour, but only during the night
redis-cli SET "brooce:cron:jobs:log-rotate" "0,15,30,45 0-8 * * * queue:common ~/bin/rotate-logs.sh"

# I have no idea why you'd want to do this
redis-cli SET "brooce:cron:jobs:log-rotate" "0-15,45-59 */3,*/4 * * * queue:common ~/bin/delete-customer-data.sh"
```

### Storing Cron Jobs in your Git Repo
We store cron jobs in redis rather than a config file because multiple brooce instances might be running on separate machines. If there was a cron.conf file, there is a risk that different versions of it might end up on the different machines.

However, nothing prevents you from creating a shell script called cron.sh that clears out and resets your cron jobs. You can then commit that script to your Git repo, and run it as part of your deploy process. It might look like this:
```shell
#!/bin/bash
redis-cli KEYS "brooce:cron:jobs:*" | xargs redis-cli DEL
redis-cli SET "brooce:cron:jobs:daily-biller" "0 0 * * * queue:common ~/bin/bill-all-accounts.sh"
redis-cli SET "brooce:cron:jobs:hourly-log-rotater" "0 * * * * queue:common ~/bin/rotate-logs.sh"
redis-cli SET "brooce:cron:jobs:twice-daily-error-checker" "0 */12 * * * queue:common ~/bin/check-for-errors.sh"
```

## Hacking on brooce
For most users, it should be enough to download our binaries. If you want to hack on the project, you should install Go and the [Gb build tool](https://getgb.io/). Then check out the repo into its own folder, and use gb to build it.
