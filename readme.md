## xtree

A k9s plugin to quickly view and navigate through `.crossplane.resourceRefs` in an XR (Composite Resource)


### Writeup

### K9s

#### Plugin development
A k9s plugin is basically just a command you decide to run, and k9s will display the output of it. If you have selected a resource k9s allows you to send this information as `arg` to the program using pre-decided variables like `$NAME` or `$NAMESPACE`.

That means you can use `kubectl` commands directy, like `kubectl get $NAME -n $NAMESPACE or accept it as flags in your custom program.

At the time of writing, the k9s `ui` code is in the `internal` directory, meaning i cant reuse it in my plugin-code for cohesity between the views.

It is a bit unfortunate, but gives me the opportunity to instead check out the [Bubbletea package](https://github.com/charmbracelet/bubbletea/) which i've seen in action in an internal tool and wanted to check out anyway.

### Bubbletea

#### ELM

### Mocks

Added mock functions that i can replace with the real operations later.
[mock.go](mock.go).

Allows me to not rely on kubernetes during development, and lets me quickly lay the foundations, find out what i need, and worry about the specifics later.


### k9s install