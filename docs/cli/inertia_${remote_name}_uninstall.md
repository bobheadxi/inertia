## inertia ${remote_name} uninstall

Shut down Inertia and remove Inertia assets from remote host

### Synopsis

Shuts down and removes the Inertia daemon, and removes the Inertia
directory (~/inertia) from your remote host.

```
inertia ${remote_name} uninstall [flags]
```

### Options

```
  -h, --help   help for uninstall
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
  -s, --short           don't stream output from command
      --verify-ssl      verify SSL communications - requires a signed SSL certificate
```

### SEE ALSO

* [inertia ${remote_name}](inertia_${remote_name}.md)	 - Configure deployment to ${remote_name}

