$env:GOOS = 'linux';
$env:GOARCH = 'arm64';

$thost = 'raspit';

go build -o linux

if (!$?)
{
	Write-Host -ForegroundColor red "ERROR: build failed"
	exit 1
}

ssh ${thost} -C "sudo service myapp stop"

#ssh ${thost} -C "rm -rf ~/app && mkdir ~/app"

scp .\linux ${thost}:~/app/app
if (!$?)
{
	Write-Host -ForegroundColor red "ERROR: scp bin"
	exit 1
}

scp .\config.linux.json ${thost}:~/app/config.json
if (!$?)
{
	Write-Host -ForegroundColor red "ERROR: scp config"
	exit 1
}

ssh ${thost} -tC "cd ~/app && chmod 0750 ./app && sudo service myapp restart"
