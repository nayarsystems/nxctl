## nxctl
[Nexus](https://github.com/jaracil/nexus/) command line tool. Wrapper of the golang [Nexus client](https://github.com/jaracil/nxcli)

## Install
    go get github.com/nayarsystems/nxctl

## Config file

The config is read using [Viper](https://github.com/spf13/viper), so it supports JSON, TOML, YAML, HCL, or Java properties formats, from any of these paths: 

* ./
* $HOME/.nxctl/
* $HOME/.config/nxctl/
* $HOME/.local/config/nxctl/
* /etc/nxctl/
* %APPDATA%/nxctl/

-- 

    $ cat /home/user/.config/nxctl/default.yml 
    user: defaultuser
    password: secretpass
    server: wss://nexus.local
    timeout: 120

    $ cat /home/user/.config/nxctl/production.json
    {
        "user" : "root" 
        "password" : "%%verysecretpass%%",
        "server" : "wss://nexus.remote.com",
    }


Config files can be selected by passing the -c flag, using the filename (without extension):

    $ ./nxctl shell 
    2016/08/12 08:21:08 Connected to wss://nexus.local
    2016/08/12 08:21:47 Logged as defaultuser
    ...

    $ ./nxctl -c production shell 
    2016/08/12 08:21:08 Connected to wss://nexus.remote.com
    2016/08/12 08:21:47 Logged as root
    ...


