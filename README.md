# brooce
Brooce is a language-agnostic job queue with a redis backend. It was written in Go.

## Features

* **Single Executable** -- Brooce comes as a single executable that runs on any Linux system.
* **Redis Backend** -- Redis can be accessed from any programming language, or the command line. Schedule jobs from anywhere.
* **Language-Agnostic** -- Jobs are just shell commands. Write jobs in any language.
* **Web Interface** -- Brooce runs its own password-protected web server. Access it to easily monitor running jobs.
* **Job Logging** -- Job stdout/stderr output can be logged to redis or log files, for later review through the web interface or your favorite text editor.
* **Scalable** -- Deploy workers on one machine or many. All workers coordinate amongst themselves.
* **Locking** -- Jobs can use brooce's lock system, or implement their own. A job that can't grab a lock it needs will be delayed and put back on the queue a minute later.
* **Cron Jobs** -- Schedule tasks to run on a schedule.
* **Suicide Mode** -- Instruct brooce to run a shell command after it's been idle for a pre-set period. Perfect for having unneeded EC2 workers terminate themselves.

## Quick Start
We have a tutorial that'll get you set up and run your first job.

[View Quick Start Tutorial](QUICKSTART.md)

 
## Configuration
The first time brooce runs, it will create a `~/.brooce` dir in your home directory with a default `~/.brooce/brooce.conf` config file. 

[View brooce.conf Documentation](CONFIG.md)
 

## Setting Up Multiple Queues
Brooce is multi-threaded, and can run many jobs at once on multiple queues.

## Cron Jobs

## Locking
