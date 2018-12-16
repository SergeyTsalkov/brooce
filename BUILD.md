# Build From Source
Now that go modules are ready, we have dropped the gb requirement. Building from source is now simple:
```shell
sudo apt-get install redis-server
git clone git@github.com:SergeyTsalkov/brooce.git brooce
cd brooce
go build
./brooce
```