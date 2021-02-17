# Simple Backblaze File Server

A simple web server that retrives & saves files in [B2 Cloud Storage](https://www.backblaze.com/b2/cloud-storage.html).

It can be accessed at http://files.127.0.0.1.nip.io:9090/ for testing in local dev environment.

Or add a mapping to the hosts file:

```bash
# /etc/hosts
127.0.0.1       files.example.com
```

Firefox DOH settings may need an update as the hosts file might be ignored when resolving via DNS over HTTPS.

```bash
# firefox: 1 - Race native against TRR.
about:config
network.trr.mode = 1
```

Overwrite the default config with these environemt variables:

```sh
export B2SERVER_BACKBLAZE_APPLICATION_KEY=
export B2SERVER_BACKBLAZE_KEY_ID=
```

Or with commandline options:

```sh
OPTIONS
  --backblaze-application-key
  --backblaze-key-id
```

