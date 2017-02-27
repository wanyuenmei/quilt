# Quilt.js
Documentation for quilt.js

## Container
The Container object represents a container to be deployed.

### Container.filepathToContent

`Container.filepathToContent` defines text files to be installed on the container
before the container starts. Both the key and value are `string`s.

For example,
```
{
  "/etc/myconf": "foo"
}
```
would create a file at `/etc/myconf` containing the text `foo`.

```
new Container("haproxy").withFiles({
  "/etc/myconf": "foo"
});
```
would create a `haproxy` instance with a text file `/etc/myconf` containing `foo`.

If the files change after the container boots, Quilt does not restart the container.
However, if the file content specified by `filepathToContent` changes, Quilt will
destroy the old container and boot a new one with the proper files.

The files are installed with permissions `0644`.
