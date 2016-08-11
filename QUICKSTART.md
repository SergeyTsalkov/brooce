# Quick Start
Just a few commands will download bruce and get it running:
```shell
sudo apt-get install redis-server
wget [url-of-binary]
./brooce
```

You'll see the output shown below:
```
Unable to read config file /home/sergey/.brooce/brooce.conf so using defaults!
You didn't specify a web username/password, so we generated these: admin/uxuavdia
We wrote a default config file to /home/sergey/.brooce/brooce.conf
Starting HTTP server on :8080
Started with queues: common (x1)
```
It's telling you that since it couldn't find your config file, it created a default one, and started the web server on port 8080. Since you haven't specified login credentials for the web interface yet, it generated some for you.

### Let's run a job!
Now open up another terminal window, and schedule your first command:
```shell
redis-cli LPUSH brooce:queue:common:pending 'ls -l ~ | tee ~/files.txt'
```

Give it a sec to run, and see that it actually ran:
```shell
cat ~/files.txt
```

### Check out the web interface!
Type `http://<yourIP>:8080` into your browser and you should see the brooce web interface come up. At the top, you'll see the "common" queue with 1 done job. Click on the hyperlinked 1 in the Done column, and you'll see some options to reschedule or delete the job. For now, just click on `Show Log` and see a listing of the files in your home directory.

### What about running jobs in parallel?
Go back to your first terminal window and hit Ctrl+C to kill brooce. Open up its config file, `~/brooce/brooce.conf` for editing. We have a [whole separate page](CONFIG.md) about all the various options, but for now, let's add another queue with 5 threads. Change the "queues" section to look like this:
```json
{
  "queues": {
    "common": 1,
    "parallel": 5
  }
}
```

Now save and re-launch brooce, and in a separate shell window, run a bunch of slow commands in our new parallel queue:
```shell
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
redis-cli LPUSH brooce:queue:parallel:pending 'sleep 30'
```
Now go back to the web interface, and note that 5 of your jobs are running, with others waiting to run. Go ahead and kill brooce again -- any jobs that are running when it dies will fail.

### Send it to the background!
Now that you're convinced that brooce is working, send it to the background:
```shell
./brooce --daemonize
```
It'll run until you kill it from the command line. Alternatively, you can use your operating system's launcher to have it run on boot.

### Lots more!
There's much more that brooce can do! Be sure to check out the [README](README.md) and the [brooce.conf documentation](CONFIG.md) for all the details!
