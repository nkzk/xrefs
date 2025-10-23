## xrefs

A k9s plugin to quickly view and navigate through `.crossplane.resourceRefs` in an XR (Composite Resource)

## Install

`go install github.com/nkzk/xrefs@latest`

## Install in k9s

`xrefs --install`

## Usage

1: Navigate to an XR in k9s
2: Press `shift+g` to show XR resource references
3: Navigate to a resource, press `y` for yaml or `d` for describe
4: Press esc/q to quit yaml or describe window, or `ctrl+c` to quit the whole program.

Some vim commands like k(up), j(down), g(top) G(bottom) are supported.

A help text is shown at the bottom for available commands.

### extra flags

Customize shortcut (have to not conflict with existing k9s shortcuts)

```sh
--shortcut "Shift+G"
```
