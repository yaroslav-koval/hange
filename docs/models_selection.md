## This documents contains explanations about reasons of using specific models.

Different commands can call different models by their reasons:

* `explain` works with one of `gpt` models, despite `gpt-codex` models would fit here better. Reason is work with files
  and vector stores, this feature is not available for `codex` models.
* `commit-msg` and `commit` commands work with `gpt-codex`, commit content is embedded into input.
