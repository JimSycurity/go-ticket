# Shell Completion Guidance

Generated shell completions are deferred until the command surface stabilizes
further. For now, users can install a small static completion list for common
shells.

## Bash

```sh
_gtk_complete() {
  local commands="help init create show edit list ls query start close reopen status dep undep link unlink ready blocked closed add-note migrate-beads super version"
  COMPREPLY=($(compgen -W "$commands" -- "${COMP_WORDS[COMP_CWORD]}"))
}
complete -F _gtk_complete gtk
```

Put that snippet in `~/.bash_completion` or a file sourced by your shell.

## zsh

```sh
#compdef gtk
_arguments "1:command:(help init create show edit list ls query start close reopen status dep undep link unlink ready blocked closed add-note migrate-beads super version)"
```

Save as a completion file on your zsh `fpath`.

## PowerShell

```powershell
Register-ArgumentCompleter -Native -CommandName gtk -ScriptBlock {
    param($wordToComplete)
    "help","init","create","show","edit","list","ls","query","start","close",
    "reopen","status","dep","undep","link","unlink","ready","blocked","closed",
    "add-note","migrate-beads","super","version" |
        Where-Object { $_ -like "$wordToComplete*" } |
        ForEach-Object { [System.Management.Automation.CompletionResult]::new($_) }
}
```

Native generated completions can replace these snippets in a later release.
