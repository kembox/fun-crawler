## Simple dotnet webapp to read fun-crawler text output from files and display it

## Environment setup
- I run it on Ubuntu Linux so we need some packages:
```shell
wget https://packages.microsoft.com/config/debian/12/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
dpkg -i packages-microsoft-prod.deb
apt-get update
apt-get install dotnet7 dotnet-runtime-7.0 dotnet-sdk-7.0 aspnetcore-runtime-7.0
```
- We'll need an nginx server in front of this.
- [fun-crawler](https://github.com/kembox/fun-crawler#quick-start) report saved in `/var/fun-crawler/reports` ( yeah I hard-coded it, sorry), named after its domain like this:
```
┌──(root㉿mox)-[/var/fun-crawler/reports]
└─# ls
tuoitre.vn  vnexpress.net
```

## How to setup

### Quick start
- Copy this folder to somewhere in your server, cd there and run:
```shell
dotnet run
```
- Setup nginx in front of it `http://127.0.0.1:5184`. For example:
```
        location / {
                proxy_set_header        Host $host:$server_port;
                proxy_set_header        X-Real-IP $remote_addr;
                proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header        X-Forwarded-Proto $scheme;
                proxy_pass http://127.0.0.1:5184;
        }
```

### From scratch

- Create dotnet webapp project from command prompt:
```
dotnet new webapp {folder_of_your_choice}
```
- Copy and place following files into your `Pages` folder.
    - [Report.cshtml](https://github.com/kembox/fun-crawler/blob/main/dotnet-webapp/top10/Pages/Report.cshtml)
    - [Report.cshtml.cs](https://github.com/kembox/fun-crawler/blob/main/dotnet-webapp/top10/Pages/Report.cshtml.cs).
    - [Index.cshtml](https://github.com/kembox/fun-crawler/blob/main/dotnet-webapp/top10/Pages/Index.cshtml)
- Run `dotnet run`. Depends on your [launchSettings](https://github.com/kembox/fun-crawler/blob/main/dotnet-webapp/top10/Properties/launchSettings.json) your app can start listening on another port than 5184.

### Logic 
Super simple. I just have a [tiny C# code](https://github.com/kembox/fun-crawler/blob/main/dotnet-webapp/top10/Pages/Report.cshtml#L3-L25) to read data from files and [generate html](https://github.com/kembox/fun-crawler/blob/main/dotnet-webapp/top10/Pages/Report.cshtml#L28-L49) for it. 

A more serious but still simple setup on server can be :
- A cron to run my crawl script and generate input to a specific folder ( with resume feature enable, every minute, with flock )
- Reconfigure this dotnet webapp to read data from there. Will need to check input more carefully. 
