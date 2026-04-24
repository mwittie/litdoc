# Hello

The following fenced code blocks should be ignored:

```bash
echo "just, bash"
```

This next one should be executed:

```bash | litdoc
echo "hello, world"
```

<!-- litdoc -->
output
<!-- /litdoc -->

Another one to be ignored:

<!--
comment
-->

And one more to be executed:

<!--bash | litdoc
echo "something to run"
-->

<!-- litdoc -->
output
<!-- /litdoc -->

Here's a previously executed block:

```bash | litdoc
echo "hello, world"
```

<!-- litdoc -->
output
<!-- /litdoc -->

And a verbatim block that should be ignored:

````md
```bash | litdoc
echo "hello, world"
```
````