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

- Fenced code block with previously generated output

```bash | litdoc
echo "hello, world"
```

<!-- BEGIN LITDOC OUTPUT -->
output
<!-- END LITDOC OUTPUT -->

- Indented code block

  ```bash | litdoc
  echo "hello, world"
  ```

- Indented code block with previously generated output

  ```bash | litdoc
  echo "hello, world"
  ```

  <!-- BEGIN LITDOC OUTPUT -->
  output
  <!-- END LITDOC OUTPUT -->

> Block quoted fenced code block with previously generated output
> 
> ```bash | litdoc
> echo "hello, world"
> ```
> 
> <!-- BEGIN LITDOC OUTPUT -->
> output
> <!-- END LITDOC OUTPUT -->

### Fenced code block indentation cases

Fenced code block indented one space:

 ```bash | litdoc
 echo "hello, world"
 ```

Fenced code block indented three spaces:

   ```bash | litdoc
   echo "hello, world"
   ```

> Block quoted fenced code block
>
> ```bash | litdoc
> echo "hello, world"
> ```

> Nested block quoted fenced code block
>
> > ```bash | litdoc
> > echo "hello, world"
> > ```

Fenced code block in an unordered list:

- ```bash | litdoc
  echo "hello, world"
  ```

Fenced code block in a nested unordered list:

  - ```bash | litdoc
    echo "hello, world"
    ```

Fenced code block in a plus list with a tilde fence:

+ ~~~bash | litdoc
  echo "hello, world"
  ~~~

Fenced code block in an ordered list:

2. ```bash | litdoc
   echo "hello, world"
   ```

Fenced code block in a nested ordered list:

   1. ```bash | litdoc
      echo "hello, world"
      ```

Fenced code block in a list blockquote:

  > ```bash | litdoc
  > echo "hello, world"
  > ```

> Fenced code block in a blockquote nested list:
>
>   - ```bash | litdoc
>     echo "hello, world"
>     ```

> Fenced code block in a blockquote ordered list:
>
> 1. ```bash | litdoc
>    echo "hello, world"
>    ```

### HTML comment indentation cases

> Block quoted HTML comment
>
> <!--bash | litdoc
> echo "hello, world"
> -->

> Nested block quoted HTML comment
>
> > <!--bash | litdoc
> > echo "hello, world"
> > -->

HTML comment in an unordered list:

- <!--bash | litdoc
  echo "hello, world"
  -->

HTML comment in a nested unordered list:

  - <!--bash | litdoc
    echo "hello, world"
    -->

HTML comment in a star list:

* <!--bash | litdoc
  echo "hello, world"
  -->

HTML comment in an ordered list:

2. <!--bash | litdoc
   echo "hello, world"
   -->

> HTML comment in a blockquote list:
>
> - <!--bash | litdoc
>   echo "hello, world"
>   -->

> HTML comment in a nested blockquote list:
>
> > - <!--bash | litdoc
> >   echo "hello, world"
> >   -->

HTML comment in an ordered-list blockquote:

   > <!--bash | litdoc
   > echo "hello, world"
   > -->
