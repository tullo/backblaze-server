# Simple Backblaze File Server

For testing local dev environment:

```bash
# /etc/hosts
127.0.0.1       file.example.com
```

Firefox DOH settings may need an update as the hosts file might be ignored.

```bash
# firefox: 1 - Race native against TRR.
about:config
network.trr.mode = 1
```
