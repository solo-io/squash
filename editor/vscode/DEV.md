# Notes for making changes to the vscode extension

## Testing download feature

### Background
- This is a way to download a particular version of squashctl from github:
```
wget https://github.com/solo-io/squash/releases/download/v0.5.3/squashctl-darwin -O latestsquash
```
- When downloading through vscode, we want the url to have the format:
```javascript
let url = `https://github.com/solo-io/squash/releases/download/${version}/squashctl-${os-name}`
```

### Setup
- since the release process writes the expected version and os-specific sha values, we need to set them explicitly, according to our development objective
- this is the json file we need to generate:
```json
{
  "version": <target_version>,
  "baseName": "squashctl",
  "binaries": {
    "win32": <associated_win_hash>,
    "linux": <associated_linux_hash>,
    "darwin": <associated_darwin_hash>
  }
}
```
### How to get the hash values
- you can download the shas like this:
```
wget https://github.com/solo-io/squash/releases/download/v0.5.3/squashctl-darwin.sha256
```
- Better, to get all shas and generate the full `editor/vscode/src/squash.json` file:
```
VERSION=<desired_released_version> make -f Makefile.dev extension-json
```
