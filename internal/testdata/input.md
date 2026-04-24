# Usage examples

## Content blocks that should be copied over without execution

- Text content (and the headings above)

- Static fenced code blocks

   ```bash
   echo "static"
   ```

- HTML comments

   <!--
   comment
   -->

- Verbatim blocks

   ````md
   ```bash | litdoc
   echo "hello, world"
   ```
   ````

## Content blocks that should be executed

- Fenced code block

```bash | litdoc
echo "hello, world"
```

- HTML comment

<!--bash | litdoc
echo "something to run"
-->
